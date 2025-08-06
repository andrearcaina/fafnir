package utils

import (
	"encoding/json"
	"errors"
	apperrors "fafnir/shared/pkg/errors"
	"log"
	"net/http"
)

// WriteJSON takes in a response writer, a status code, and data to write as JSON
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		HandleError(w, err)
		return
	}
}

// DecodeJSON reads JSON from the request body and returns a proper AppError for validation failures
func DecodeJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return apperrors.ValidationError("Request body is required")
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("Failed to close request body: %v", err)
		}
	}()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return apperrors.ValidationError("Invalid JSON format in request body")
	}

	return nil
}

// HandleError is the central errors handler for HTTP responses
func HandleError(w http.ResponseWriter, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		if appErr.Cause != nil {
			log.Printf("AppError [%s]: %s (caused by: %v)", appErr.Code, appErr.Message, appErr.Cause)
		} else {
			log.Printf("AppError [%s]: %s", appErr.Code, appErr.Message)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.HTTPStatus)
		w.Write(appErr.ToJSON())
		return
	}
}
