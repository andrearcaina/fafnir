package utils

import (
	"encoding/json"
	"net/http"
)

// WriteJSON takes in a response writer, a status code, and data to write as JSON
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
	return nil
}

// ParseJSON takes in an HTTP request and an interface to decode the JSON body into
func ParseJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
