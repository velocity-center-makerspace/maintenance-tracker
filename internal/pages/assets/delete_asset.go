package assets

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"time"

	"github.com/velocity-center-makerspace/maintenance-tracker/internal/response"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/router"
)

type DeleteAssetRequest struct {
	AssetID string `json:"asset_id"`
}

func RegisterDeleteAsset(r router.Router) {
	r.AddRoute("DELETE /assets", DeleteAsset)
}

func DeleteAsset(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {
	var req DeleteAssetRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		resp := response.New("Unknown fields in request body")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Unknown fields in DeleteAsset request body", "error", err)
		return
	}

	if req.AssetID == "" {
		resp := response.New("No ID in request body")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("No AssetID passed into DeleteAsset request body")
		return
	}

	assetID, err := uuid.Parse(req.AssetID)
	if err != nil {
		resp := response.New("Invalid id passed in request")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client did not pass valid UUID in request", "error", err)
		return
	}

	tx, err := deps.DB.BeginTx(r.Context(), nil)
	if err != nil {
		resp := response.New("Something went wrong; asset could not be deleted")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error("Could not start transaction for asset deletion", "error", err)
		return
	}

	defer func() {
		for attempts := range 3 {
			if err := tx.Rollback(); err != nil {
				if errors.Is(err, sql.ErrTxDone) {
					return
				}

				// incrementally longer retrys
				time.Sleep(10 * (1 << attempts) * time.Millisecond)
			}
			slog.Error("Failed to rollback", "error", err)
		}
	}()

	qtx := deps.Qry.WithTx(tx)

	rows, err := qtx.DeleteAssetByID(r.Context(), assetID)
	if err != nil || rows < 1 {
		resp := response.New("Something went wrong; asset could not be deleted")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error(
			"Database transaction failed; could not delete asset",
			"error",
			err,
			"rows-affected",
			rows,
		)
		return
	}

	contentHashes, err := qtx.ReadContentHashByAssetID(r.Context(), assetID)
	if err != nil {
		resp := response.New("Something went wrong; asset could not be deleted")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error(
			"Database transaction failed; could not retrieve content hashes",
			"error",
			err,
		)
		return
	}

	for _, hash := range contentHashes {
		rows, err := qtx.DeleteFileByContentHash(r.Context(), hash)
		if err != nil {
			resp := response.New("Something went wrong; asset could not be deleted")
			resp.Write(w, http.StatusInternalServerError)
			slog.Error(
				"Database transaction failed; could not delete files",
				"error",
				err,
				"rows-affected",
				rows,
			)
			return
		}
	}

	rows, err = qtx.DeleteAssetFileByAssetID(r.Context(), assetID)
	if err != nil {
		resp := response.New("Something went wrong; asset could not be deleted")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error(
			"Database transaction failed; could not delete asset files",
			"error",
			err,
			"rows-affected",
			rows,
		)
		return
	}

	for attempts := range 3 {
		err = tx.Commit()
		if err == nil {
			break
		}

		// incrementally longer retrys
		time.Sleep(10 * (1 << attempts) * time.Millisecond)
	}

	if err != nil {
		resp := response.New("Something went wrong; asset could not be deleted")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error(
			"Database transaction failed; could not commit transaction",
			"error",
			err,
		)
		return
	}

	resp := response.New("Deleted Asset successfully")
	resp.ID = assetID.String()
	resp.Write(w, http.StatusNoContent)
}
