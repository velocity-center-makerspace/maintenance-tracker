package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-sqlite3"

	"github.com/google/uuid"
	"github.com/velocity-center-makerspace/maintenance-tracker/db"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/server"
)

type upload struct {
	tee      io.Reader // created from io.TeeReader(r, upload.hash)
	hash     hash.Hash
	ext      string
	filename string
	mimeType string
}

type fileMetaData struct {
	contentHash string
	mimeType    string
	filename    string
	path        string
}

func AddAssetHandlers(mux *server.Mux) {
	mux.HandleFunc("POST /assets", func(w http.ResponseWriter, r *http.Request) {
		createAssetWithFileUpload(mux, w, r)
	})
}

func createAssetWithFileUpload(mux *server.Mux, w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, "Unable to retrieve form data", http.StatusInternalServerError)
		slog.Error("Unable to parse mime type", "error", err)
		return
	}

	if !(strings.HasPrefix(mediaType, "multipart/")) && r.Method != "POST" {
		http.Error(w, "Incorrect mime type or request method", http.StatusBadRequest)
		return
	}

	var asset db.CreateAssetParams
	var uploads []upload

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Unable to retrieve form data", http.StatusInternalServerError)
		slog.Error("Unable to read request", "error", err)
		return
	}

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			http.Error(w, "Unable to retrieve form data", http.StatusInternalServerError)
			slog.Error("Unable to read next form part", "error", err)
			return
		}

		mimeType := part.Header.Get("Content-Type")

		switch mimeType {
		case "application/json":
			if err := json.NewDecoder(part).Decode(&asset); err != nil {
				http.Error(w, "Unable to retrieve form data", http.StatusInternalServerError)
				slog.Error("Unable to decode JSON", "error", err)
				return
			}
			asset.ID, err = uuid.NewV7()
			if err != nil || asset.ID == uuid.Nil {
				http.Error(w, "Unable to retrieve form data", http.StatusInternalServerError)
				slog.Error("Unable to set Asset ID", "error", err, "asset_id", asset.ID)
				return
			}
		case "application/pdf":
			fallthrough
		case "application/msword":
			fallthrough
		case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
			if part.FileName() == "" {
				http.Error(w, "No file name found", http.StatusBadRequest)
				return
			}
			// TODO: fix tee overwrite bug
			hash := sha256.New()
			tee := io.TeeReader(part, hash)

			uploads = append(uploads,
				upload{
					tee:      tee,
					hash:     hash,
					mimeType: mimeType,
					filename: part.FileName(),
					ext:      filepath.Ext(part.FileName()),
				},
			)
		}
	}

	errChan := make(chan error, len(uploads))

	wg := sync.WaitGroup{}
	wg.Add(len(uploads))

	dstRoot := mux.Env.UploadRoot

	metaDatas := []fileMetaData{}
	metaDataMut := sync.Mutex{}

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
				metaDataMut.Lock()
				metaDatas = append(metaDatas, *meta)
				metaDataMut.Unlock()
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
		http.Error(w, "File upload failed", http.StatusInternalServerError)
		return
	}

	tx, err := mux.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(
			w,
			"File upload failed",
			http.StatusInternalServerError,
		)
		slog.Error("Failed to initiate transaction", "error", err)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			slog.Error("Failed to rollback", "error", err)
		}
	}()

	qtx := mux.Qry.WithTx(tx)

	rows, err := qtx.CreateAsset(r.Context(), asset)
	if err != nil {
		http.Error(
			w,
			"File upload failed",
			http.StatusInternalServerError,
		)
		slog.Error("Database insert failed for asset", "error", err, "rows-affected", rows)
		return
	}

	for _, m := range metaDatas {
		fileParams := db.CreateFileParams{
			ContentHash: m.contentHash,
			MimeType:    m.mimeType,
			Path:        m.path,
		}

		assetFileParams := db.CreateAssetFileParams{
			ContentHash:      m.contentHash,
			AssetID:          asset.ID,
			OriginalFilename: m.filename,
		}

		rows, err := qtx.CreateFile(r.Context(), fileParams)
		if err != nil || rows < 1 {
			if isDuplicateKeyErr(err) {
				continue
			}
			http.Error(
				w,
				"File upload failed",
				http.StatusInternalServerError,
			)
			slog.Error("Database insert failed for file", "error", err, "rows-affected", rows)
			return
		}

		rows, err = qtx.CreateAssetFile(r.Context(), assetFileParams)

		if err != nil || rows < 1 {
			if isDuplicateKeyErr(err) {
				continue
			}

			http.Error(
				w,
				"File upload failed",
				http.StatusInternalServerError,
			)
			slog.Error("Database insert failed for asset file", "error", err, "rows-affected", rows)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Unable to create asset", http.StatusInternalServerError)
		slog.Error("Transaction commit failed", "error", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

type processFileParams struct {
	upload  upload
	dstRoot string
	errChan chan<- error
}

func processFile(p processFileParams) *fileMetaData {
	tmp, err := os.CreateTemp(p.dstRoot, "upload-*.tmp")
	if err != nil {
		p.errChan <- err
		return nil
	}

	defer func() {
		for attempts := range 3 {
			err := os.Remove(tmp.Name())
			if err == nil || errors.Is(err, os.ErrNotExist) {
				return
			}

			// incrementally longer retrys
			time.Sleep(10 * (1 << attempts) * time.Millisecond)
		}
		p.errChan <- fmt.Errorf("failed to remove after 3 attempts: %w", err)
	}()

	defer func() { _ = tmp.Close() }() // not handling err b/c will be removed regardless

	if _, err = io.Copy(tmp, p.upload.tee); err != nil {
		p.errChan <- err
		return nil
	}

	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		p.errChan <- err
		return nil
	}

	checksum := hex.EncodeToString(p.upload.hash.Sum(nil))
	dstPath := filepath.Join(
		p.dstRoot,
		checksum[:2],
		checksum[2:4],
		checksum+p.upload.ext,
	)

	if _, err = os.Stat(dstPath); err == nil {
		return &fileMetaData{
			contentHash: checksum,
			filename:    p.upload.filename,
			mimeType:    p.upload.mimeType,
			path:        dstPath,
		}
	}

	if err = os.MkdirAll(filepath.Dir(dstPath), 0744); err != nil {
		p.errChan <- err
		return nil
	}

	// create dst & copy temp into it
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

	return &fileMetaData{
		contentHash: checksum,
		filename:    p.upload.filename,
		mimeType:    p.upload.mimeType,
		path:        dstPath,
	}
}

func isDuplicateKeyErr(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrConstraint
	}
	return false
}
