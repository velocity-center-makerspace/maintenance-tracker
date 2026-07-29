package assets

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-sqlite3"

	"github.com/google/uuid"
	"github.com/velocity-center-makerspace/maintenance-tracker/db"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/response"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/router"
)

var ErrUnsupportedContentType = errors.New("unsupported content type")

type upload struct {
	tmpPath  string
	checksum string
	ext      string
	filename string
	mimeType string
}

type fileMetadata struct {
	contentHash string
	mimeType    string
	filename    string
	path        string
}

func RegisterCreateAsset(r router.Router) {
	r.AddRoute("POST /assets", CreateAsset)
}

func CreateAsset(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		resp := response.New("Unable to retrieve form data")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error("Unable to parse mime type", "error", err)
		return
	}

	if !(strings.HasPrefix(mediaType, "multipart/")) && r.Method != "POST" {
		resp := response.New("Incorrect mime type or request method")
		resp.Write(w, http.StatusBadRequest)
		slog.Warn("Client sent a bad request", "mediaType", mediaType)
		return
	}

	asset := db.CreateAssetParams{}

	dstRoot := deps.Env.UploadRoot

	reader, err := r.MultipartReader()
	if err != nil {
		resp := response.New("Unable to retrieve form data")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error("Unable to read request", "error", err)
		return
	}

	params := parsePartsParams{
		asset:   &asset,
		reader:  reader,
		dstRoot: dstRoot,
	}

	uploads, err := parseParts(params)
	if err != nil {
		if errors.Is(err, ErrUnsupportedContentType) {
			resp := response.New("Incompatible file type submitted")
			resp.Write(w, http.StatusBadRequest)
			slog.Warn("Client submitted incompatible file type", "error", err)
			return
		}
		resp := response.New("Unable to retrieve form data")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error("Unable to parse multipart reader parts", "error", err)
		return
	}

	errChan := make(chan error, len(uploads))

	wg := sync.WaitGroup{}
	wg.Add(len(uploads))

	metadatas := []fileMetadata{}
	metadataMut := sync.Mutex{}

	for _, u := range uploads {
		go func(u upload) {
			defer wg.Done()

			params := processFileParams{
				upload:  u,
				dstRoot: dstRoot,
				errChan: errChan,
			}

			meta := processFile(params)

			if meta != nil {
				metadataMut.Lock()
				metadatas = append(metadatas, *meta)
				metadataMut.Unlock()
			}
		}(u)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var firstErr error
	for err := range errChan {
		slog.Error("Unable to process file: ", "error", err)
		if firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		resp := response.New("File upload failed")
		resp.Write(w, http.StatusInternalServerError)
		return
	}

	txParams := assetTxParams{
		db:        deps.DB,
		ctx:       r.Context(),
		metadatas: metadatas,
	}

	err = assetTx(txParams)
	if err != nil {
		resp := response.New("Unable to save asset. Please try again later.")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error("Database transaction failed", "error", err)
		return
	}

	resp := response.New("Asset created successfully")
	resp.ID = asset.ID.String()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	w.WriteHeader(http.StatusCreated)
}

type parsePartsParams struct {
	reader  *multipart.Reader
	asset   *db.CreateAssetParams
	dstRoot string
}

func parseParts(p parsePartsParams) ([]upload, error) {
	var uploads []upload
	for {
		part, err := p.reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		mimeType := part.Header.Get("Content-Type")

		switch mimeType {
		default:
			return nil, fmt.Errorf(
				"%w: %q for part %s",
				ErrUnsupportedContentType,
				mimeType,
				part.FormName(),
			)
		case "application/json":
			if err := json.NewDecoder(part).Decode(p.asset); err != nil {
				return nil, err
			}
			p.asset.ID, err = uuid.NewV7()
			if err != nil || p.asset.ID == uuid.Nil {
				return nil, fmt.Errorf("asset ID is %v and could not be set: %w", p.asset.ID, err)
			}
		case "application/pdf":
			fallthrough
		case "application/msword":
			fallthrough
		case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
			if part.FileName() == "" {
				continue
			}

			tmp, err := os.CreateTemp(p.dstRoot, "upload-*.tmp")
			if err != nil {
				return nil, err
			}

			h := sha256.New()
			tee := io.TeeReader(part, h)

			if _, err := io.Copy(tmp, tee); err != nil {
				return nil, err
			}

			if err := tmp.Close(); err != nil {
				return nil, err
			}

			uploads = append(uploads,
				upload{
					tmpPath:  tmp.Name(),
					checksum: hex.EncodeToString(h.Sum(nil)),
					mimeType: mimeType,
					filename: part.FileName(),
					ext:      filepath.Ext(part.FileName()),
				},
			)
		}
	}

	return uploads, nil
}

type processFileParams struct {
	upload  upload
	dstRoot string
	errChan chan<- error
}

func processFile(p processFileParams) *fileMetadata {
	tmp, err := os.Open(p.upload.tmpPath)
	if err != nil {
		p.errChan <- err
		return nil
	}

	defer func() {
		for attempts := range 3 {
			err := os.Remove(p.upload.tmpPath)
			if err == nil || errors.Is(err, os.ErrNotExist) {
				return
			}

			// incrementally longer retrys
			time.Sleep(10 * (1 << attempts) * time.Millisecond)
		}
		p.errChan <- fmt.Errorf("failed to remove after 3 attempts: %w", err)
	}()

	defer func() { _ = tmp.Close() }()

	dstPath := filepath.Join(
		p.dstRoot,
		p.upload.checksum[:2],
		p.upload.checksum[2:4],
		p.upload.checksum+p.upload.ext,
	)

	if _, err := os.Stat(dstPath); err == nil {
		return &fileMetadata{
			contentHash: p.upload.checksum,
			filename:    p.upload.filename,
			mimeType:    p.upload.mimeType,
			path:        dstPath,
		}
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0744); err != nil {
		p.errChan <- err
		return nil
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		p.errChan <- err
		return nil
	}

	defer func() {
		if err := dst.Close(); err != nil {
			p.errChan <- err
		}
	}()

	if _, err := io.Copy(dst, tmp); err != nil {
		p.errChan <- err
		return nil
	}

	return &fileMetadata{
		contentHash: p.upload.checksum,
		filename:    p.upload.filename,
		mimeType:    p.upload.mimeType,
		path:        dstPath,
	}
}

type assetTxParams struct {
	db        *sql.DB
	ctx       context.Context
	metadatas []fileMetadata
	qry       *db.Queries
	asset     db.CreateAssetParams
}

func assetTx(a assetTxParams) error {
	tx, err := a.db.BeginTx(a.ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		for attempts := range 3 {
			if err := tx.Rollback(); err != nil {
				if errors.Is(err, sql.ErrTxDone) {
					return
				}

				slog.Error("Failed to rollback", "error", err)
				// incrementally longer retrys
				time.Sleep(10 * (1 << attempts) * time.Millisecond)
			}
		}
	}()

	qtx := a.qry.WithTx(tx)

	rows, err := qtx.CreateAsset(a.ctx, a.asset)
	if err != nil {
		return fmt.Errorf(
			"database transaction failed with %d rows affected: %w",
			rows,
			err,
		)
	}

	for _, m := range a.metadatas {
		fileParams := db.CreateFileParams{
			ContentHash: m.contentHash,
			MimeType:    m.mimeType,
			Path:        m.path,
		}

		assetFileParams := db.CreateAssetFileParams{
			ContentHash:      m.contentHash,
			AssetID:          a.asset.ID,
			OriginalFilename: m.filename,
		}

		rows, err := qtx.CreateFile(a.ctx, fileParams)
		if err != nil {
			if isDuplicateKeyErr(err) {
				continue
			}
			return fmt.Errorf(
				"database transaction failed with %d rows affected: %w",
				rows,
				err,
			)
		}

		rows, err = qtx.CreateAssetFile(a.ctx, assetFileParams)

		if err != nil {
			if isDuplicateKeyErr(err) {
				continue
			}

			return fmt.Errorf(
				"database transaction failed with %d rows affected: %w",
				rows,
				err,
			)
		}
	}

	for attempts := range 3 {
		err := tx.Commit()
		if err == nil {
			return nil
		}

		// incrementally longer retrys
		time.Sleep(10 * (1 << attempts) * time.Millisecond)
	}
	return fmt.Errorf("failed to commit transaction: %w", err)
}

func isDuplicateKeyErr(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrConstraint
	}
	return false
}
