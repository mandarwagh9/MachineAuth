package handlers

import (
	"encoding/json"
	"net/http"

	"machineauth/internal/models"
)

// writeJSON sets Content-Type: application/json and encodes v to w.
// The HTTP status defaults to 200 OK (caller must not have called WriteHeader yet).
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// writeJSONStatus sets Content-Type: application/json, writes the given HTTP
// status code, and encodes v to w.
func writeJSONStatus(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeJSONError writes an RFC-6749-style JSON error response.
//
//	errCode    – the short machine-readable error token (e.g. "invalid_request")
//	description – human-readable detail
func writeJSONError(w http.ResponseWriter, status int, errCode, description string) {
	writeJSONStatus(w, status, models.ErrorResponse{
		Error:            errCode,
		ErrorDescription: description,
	})
}
