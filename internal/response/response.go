package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Response struct {
	ID  string `json:"id"`
	Msg string `json:"msg"`
}

func New(msg string) *Response {
	return &Response{
		Msg: msg,
	}
}

func (r *Response) Write(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(r); err != nil {
		http.Error(w, "Unable to send error response", http.StatusInternalServerError)
		slog.Error("Unable to send error response", "error", err)
	}
	w.WriteHeader(status)
}
