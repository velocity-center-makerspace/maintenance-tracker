package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Response struct {
	Msg  string `json:"msg"`
	Args map[string]any
}

func New(msg string) *Response {
	return &Response{
		Msg: msg,
	}
}

func NewWithArgs(msg string, args map[string]any) *Response {
	return &Response{
		Msg:  msg,
		Args: args,
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
