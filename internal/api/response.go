package api

import (
	"encoding/json"
	"net/http"
)

// writeJSON writes payload as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// errorResponse is the JSON body returned for all error conditions.
type errorResponse struct {
	Error string `json:"error"`
}

// writeError writes a JSON error body with the given status code.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
