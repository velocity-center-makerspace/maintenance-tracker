package pages

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/mattn/go-sqlite3"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/velocity-center-makerspace/maintenance-tracker/db"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/server"
)

func AddAssetHandlers(mux *server.Mux) {
	mux.HandleFunc("POST /assets", func(w http.ResponseWriter, r *http.Request) {
		createAsset(mux, w, r)
	})
}

func createAsset(mux *server.Mux, w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, "Unable to retrieve form data", http.StatusInternalServerError)
		slog.Error("Unable to parse mime type", "error", err)
		return
	}

	if !(strings.HasPrefix(mediaType, "multipart/") && r.Method == "POST") {
		http.Error(w, "Incorrect mime type or request method", http.StatusBadRequest)
		return
	}

	var asset db.CreateAssetParams
	var assetFiles []db.CreateAssetFileParams

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

			dstRoot := mux.Env.UploadRoot
			temp, err := os.CreateTemp(dstRoot, "upload-*.tmp")
			if err != nil {
				http.Error(
					w,
					"Unable to retrieve form data",
					http.StatusInternalServerError,
				)
				slog.Error("Unable to create temporary file", "error", err)
				return
			}

			defer func() {
				if err := os.Remove(temp.Name()); err != nil {
					http.Error(
						w,
						"Unable to retrieve form data",
						http.StatusInternalServerError,
					)
					slog.Error("Unable to remove temporary file", "error", err)
				}
			}()

			defer func() {
				if err := temp.Close(); err != nil {
					http.Error(
						w,
						"Unable to retrieve form data",
						http.StatusInternalServerError,
					)
					slog.Error("Unable to close temporary file", "error", err)
				}
			}()

			h := sha256.New()
			tee := io.TeeReader(part, h)

			if _, err := io.Copy(temp, tee); err != nil {
				http.Error(
					w,
					"Unable to retrieve form data",
					http.StatusInternalServerError,
				)
				slog.Error("Unable to read part file into temp file", "error", err)
				return
			}

			if _, err := temp.Seek(0, io.SeekStart); err != nil {
				http.Error(
					w,
					"Unable to retrieve form data",
					http.StatusInternalServerError,
				)
				slog.Error("Unable to move temp file cursor to start", "error", err)
				return
			}

			checksum := hex.EncodeToString(h.Sum(nil))
			fileType := filepath.Ext(part.FileName())
			dstPath := filepath.Join(
				dstRoot,
				checksum[:2],
				checksum[2:4],
				checksum+fileType,
			)
			fileExists := false

			if _, err := os.Stat(dstPath); err == nil {
				fileExists = true
			}

			if fileExists {
				continue
			}
			if err := os.MkdirAll(filepath.Dir(dstPath), 0744); err != nil {
				http.Error(
					w,
					"Unable to retrieve form data",
					http.StatusInternalServerError,
				)
				slog.Error("Unable to create directories for dst file", "error", err)
				return
			}

			dst, err := os.Create(dstPath)
			if err != nil {
				http.Error(
					w,
					"Unable to retrieve form data",
					http.StatusInternalServerError,
				)
				slog.Error("Unable to create dst file", "error", err)
				return
			}

			defer func() {
				if err := dst.Close(); err != nil {
					slog.Error("Unable to close dst file", "error", err)
				}
			}()

			if _, err = io.Copy(dst, temp); err != nil {
				http.Error(
					w,
					"Unable to retrieve form data",
					http.StatusInternalServerError,
				)
				slog.Error("Unable to copy temp file to dst file", "error", err)
				return
			}

			assetFiles = append(
				assetFiles,
				db.CreateAssetFileParams{
					ContentHash:      checksum,
					MimeType:         mimeType,
					OriginalFilename: part.FileName(),
				},
			)
		}
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
	defer tx.Rollback()

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

	for _, f := range assetFiles {
		f.AssetID = asset.ID
		rows, err := qtx.CreateAssetFile(r.Context(), f)

		if err != nil || rows < 1 {
			// TODO: need to update tables to allow for one file to potentially apply to multiple assets
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

func isDuplicateKeyErr(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrConstraint
	}
	return false
}

func readAssetByID(w http.ResponseWriter, r *http.Request) {

}

func updateAssetByID(w http.ResponseWriter, r *http.Request) {

}

func updateAssetEOLByID(w http.ResponseWriter, r *http.Request) {

}

func updateAssetFileByID(w http.ResponseWriter, r *http.Request) {

}

func updateAssetNameByID(w http.ResponseWriter, r *http.Request) {

}

func updateAssetStatusByID(w http.ResponseWriter, r *http.Request) {

}

func updateAssetWarrantyByID(w http.ResponseWriter, r *http.Request) {

}
