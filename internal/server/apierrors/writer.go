package apierrors

import (
	"encoding/json"
	"net/http"
)

// WriteJSONError writes a JSON error body with at least `{"error": "..."}`.
func WriteJSONError(w http.ResponseWriter, status int, apiErr APIError) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiErr)
}
