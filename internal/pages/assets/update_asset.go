package assets

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"strings"
	"time"

	"github.com/google/uuid"

	"net/http"

	"github.com/velocity-center-makerspace/maintenance-tracker/db"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/response"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/router"
)

func RegisterUpdateAsset(r router.Router) {
	r.AddRoute("PATCH /assets/name", UpdateAssetName)
	r.AddRoute("PATCH /assets/warranty-expiry", UpdateAssetWarrantyExpiry)
	r.AddRoute("PATCH /assets/status", UpdateAssetStatus)
	r.AddRoute("PATCH /assets/end-of-life", UpdateAssetEOL)
	r.AddRoute("PUT /assets", UpdateAsset)
}

type assetNameRequest struct {
	AssetID string `json:"asset_id"`
	Name    string `json:"asset_name"`
}

func UpdateAssetName(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {
	var req assetNameRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		resp := response.New("Request has unknown fields")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent request w/ disallowed fields")
		return
	}

	if req.AssetID == "" || req.Name == "" {
		resp := response.New("Request missing required fields")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent request with missing fields")
		return
	}

	assetID, err := uuid.Parse(req.AssetID)
	if err != nil {
		resp := response.New("Invalid id passed in request")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client did not pass valid UUID in request", "error", err)
		return
	}

	rows, err := deps.Qry.UpdateAssetNameByID(r.Context(), req.Name, assetID)
	if err != nil || rows < 1 || rows > 1 {
		resp := response.New("Something went wrong; unable to update asset name")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error(
			"Could not update asset name",
			"error",
			err,
			"asset-id",
			req.AssetID,
			"rows-affected",
			rows,
		)
		return
	}

	resp := response.New("Asset name updated successfully")
	resp.Write(w, http.StatusOK)
}

type assetWarrantyRequest struct {
	AssetID        string `json:"asset_id"`
	WarrantyExpiry string `json:"warranty_expiry"`
}

func UpdateAssetWarrantyExpiry(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {
	var req assetWarrantyRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		resp := response.New("Request has unknown fields")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent request w/ disallowed fields")
		return
	}

	if req.AssetID == "" {
		resp := response.New("Request missing required fields")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent request with missing fields")
		return
	}

	assetID, err := uuid.Parse(req.AssetID)
	if err != nil {
		resp := response.New("Invalid id passed in request")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client did not pass valid UUID in request", "error", err)
		return
	}

	var warrantyExpiry sql.NullTime
	if req.WarrantyExpiry != "" {
		timeVal, err := time.Parse(time.RFC3339, req.WarrantyExpiry)
		if err != nil {
			resp := response.New("Invalid time layout passed in request")
			resp.Write(w, http.StatusBadRequest)
			slog.Error("Client did not pass a valid warranty time layout in request", "error", err)
			return
		}

		warrantyExpiry.Time = timeVal
		warrantyExpiry.Valid = true
	} else {
		warrantyExpiry.Valid = false
	}

	rows, err := deps.Qry.UpdateAssetWarrantyByID(r.Context(), warrantyExpiry, assetID)
	if err != nil || rows < 1 || rows > 1 {
		resp := response.New("Something went wrong; unable to update warranty expiry")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error(
			"Failed to update warranty expiry",
			"error",
			err,
			"asset-id",
			req.AssetID,
			"rows-affected",
			rows,
		)
		return
	}

	resp := response.New("Asset warranty expiry was updated successfully")
	resp.Write(w, http.StatusOK)
}

type assetStatusRequest struct {
	AssetID         string `json:"asset_id"`
	Availability    string `json:"availabilty"`
	AttentionNeeded string `json:"attention_needed"`
}

func UpdateAssetStatus(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {
	var req assetStatusRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		resp := response.New("Request has unknown fields")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent request w/ disallowed fields")
		return
	}

	if req.AssetID == "" || req.Availability == "" || req.AttentionNeeded == "" {
		resp := response.New("Request missing required fields")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent request with missing fields")
		return
	}

	assetID, err := uuid.Parse(req.AssetID)
	if err != nil {
		resp := response.New("Invalid id passed in request")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client did not pass valid UUID in request", "error", err)
		return
	}

	err = checkAvailability(req.Availability)
	if err != nil {
		resp := response.New("Invalid availability string in request")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent invalid availability string", "error", err)
		return
	}

	err = checkAttentionNeeded(req.AttentionNeeded)
	if err != nil {
		resp := response.New("Invalid attention_needed in request")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent invalid attention_needed string", "error", err)
		return
	}

	err = validateAssetStatus(req.Availability, req.AttentionNeeded)
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

	params := db.UpdateAssetStatusByIDParams{
		AttentionNeeded: req.AttentionNeeded,
		Availability:    req.Availability,
		ID:              assetID,
	}
	rows, err := deps.Qry.UpdateAssetStatusByID(r.Context(), params)
	if err != nil || rows < 1 {
		resp := response.New("Something went wrong; unable to update asset status")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error("Unable to update asset status", "error", err)
		return
	}

	resp := response.New("Asset status updated successfully")
	resp.Write(w, http.StatusOK)
}

type assetEOLRequest struct {
	AssetID   string `json:"asset_id"`
	EndOfLife string `json:"end_of_life"`
}

func UpdateAssetEOL(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {
	var req assetEOLRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		resp := response.New("Request has unknown fields")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent request w/ disallowed fields")
		return
	}

	if req.AssetID == "" {
		resp := response.New("Request missing required fields")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent request with missing fields")
		return
	}

	assetID, err := uuid.Parse(req.AssetID)
	if err != nil {
		resp := response.New("Invalid id passed in request")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client did not pass valid UUID in request", "error", err)
		return
	}

	var endOfLife sql.NullTime
	if req.EndOfLife != "" {
		timeVal, err := time.Parse(time.RFC3339, req.EndOfLife)
		if err != nil {
			resp := response.New("Invalid end_of_life date passed in request")
			resp.Write(w, http.StatusBadRequest)
			slog.Error("Client did not pass valid end_of_life date in request", "error", err)
		}

		endOfLife.Time = timeVal
		endOfLife.Valid = true
	} else {
		endOfLife.Valid = false
	}

	rows, err := deps.Qry.UpdateAssetEOLByID(r.Context(), endOfLife, assetID)
	if err != nil || rows < 1 {
		resp := response.New("Something went wrong; unable to update end_of_life")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error("Unable to update end_of_life", "error", err, "rows-affected", rows)
		return
	}

	resp := response.New("Asset end-of-life updated successfully")
	resp.Write(w, http.StatusOK)
}

type removeAssetFileRequest struct {
}

func UpdateAssetFile(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		resp := response.New("Unable to retrieve form data")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error("Unable to parse mime type", "error", err)
		return
	}

	if !(strings.HasPrefix(mediaType, "multipart/")) && r.Method != "PATCH" {
		resp := response.New("Unable to retrieve updated files")
		resp.Write(w, http.StatusBadRequest)
		slog.Error("Client sent not multipart request for file updates", "mimeType", mediaType)
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		resp := response.New("Unable to retrieve updated files")
		resp.Write(w, http.StatusInternalServerError)
		slog.Error("Could not create multipart reader", "error", err)
		return
	}

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			resp := response.New("Unable to retrieve updated files")
			resp.Write(w, http.StatusInternalServerError)
			slog.Error("Could not get next part from reader", "error", err)
			return
		}

		mimeType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			resp := response.New("Unable to retrieve updated files")
			resp.Write(w, http.StatusInternalServerError)
			slog.Error("Could not parse mime type from part", "error", err)
			return
		}

		switch mimeType {
		case "application/pdf":
			fallthrough
		case "application/msword":
			fallthrough
		case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		}
	}
}

func UpdateAsset(deps router.Dependencies, w http.ResponseWriter, r *http.Request) {

}
