package utils

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// WriteJSON takes in a response writer, a status code, and data to write as JSON
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}

// DecodeJSON reads JSON from the request body (r) into v. It returns an error if decoding fails.
func DecodeJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("Failed to close request body: %v", err)
		}
	}()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errors.New("failed to decode request body")
	}

	return nil
}

// WriteError writes an error response in JSON format and logs the error
func WriteError(w http.ResponseWriter, status int, data interface{}, err error) {
	log.Printf("Error: %v", err)
	WriteJSON(w, status, data)
}
