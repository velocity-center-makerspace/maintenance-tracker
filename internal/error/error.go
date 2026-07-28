package error

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Msg string `json:"msg"`
}

func New(msg string) *ErrorResponse {
	return &ErrorResponse{Msg: msg}
}

func (e *ErrorResponse) Write(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/asset-manager")
	if err := json.NewEncoder(w).Encode(e); err != nil {
		http.Error(w, "Unable to send error response", http.StatusInternalServerError)
		slog.Error("Unable to send error response", "error", err)
	}
	w.WriteHeader(status)
}
