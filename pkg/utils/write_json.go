package utils

import (
	"encoding/json"
	"net/http"
)

// WriteJSON function that writes a JSON response to the http.ResponseWriter (doesn't handle errors, just writes the response)
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
