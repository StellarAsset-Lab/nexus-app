package httpapi

import (
	"encoding/json"
	"net/http"
)

// ErrorBody is the stable error shape every handler error uses. Never put a
// raw SQL error, stack trace, or internal path in Message — see spec §14.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error ErrorBody `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: ErrorBody{Code: code, Message: message}})
}
