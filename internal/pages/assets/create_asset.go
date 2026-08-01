package assets

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
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

	tempRoot := deps.Env.TempUploadRoot

	reader, err := r.MultipartReader()
	if err != nil {
		resp := response.New("Unable to retrieve form data")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error("Unable to read request", "error", err)
		return
	}

	uploads, dec, err := parseMultipart(reader, tempRoot)
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

	if err := dec.Decode(&asset); err != nil {
		resp := response.New("Unable to read JSON from request")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error("Unable to read JSON from request", "error", err)
		return
	}

	asset.ID, err = uuid.NewV7()
	if err != nil {
		resp := response.New("Unable to create asset from data")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error("Unable to create ID for new asset", "error", err, "asset-id", asset.ID)
		return
	}

	err = checkAvailability(asset.Availability)
	if err != nil {
		resp := response.New("Invalid availability string in request")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent invalid availability string", "error", err)
		return
	}

	err = checkAttentionNeeded(asset.AttentionNeeded)
	if err != nil {
		resp := response.New("Invalid attention_needed in request")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent invalid attention_needed string", "error", err)
		return
	}

	err = validateAssetStatus(asset.Availability, asset.AttentionNeeded)
	if err != nil {
		resp := response.New("Invalid availability and attention_needed status combination")
		resp.Write(w, http.StatusBadRequest)
		slog.Error(
			"Client sent invalid availability and attention_needed combination",
			"error",
			err,
		)
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
				dstRoot: deps.Env.UploadRoot,
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

	var args map[string]any
	args["asset_id"] = asset.ID
	resp := response.NewWithArgs("Asset created successfully", args)
	resp.Write(w, http.StatusCreated)
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
