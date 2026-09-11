// Package rest holds the HTTP controllers.
package rest

import (
	"encoding/json"
	"net/http"
)

// writeJSON renders a value as the response body with the given status.
func writeJSON(w http.ResponseWriter, status int, body any) {
	encoded, err := json.Marshal(body)
	if err != nil {
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(encoded)
}

// writeEmptyOK answers 200 with no body, which is how the original renders a
// handler that returned no object.
func writeEmptyOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
