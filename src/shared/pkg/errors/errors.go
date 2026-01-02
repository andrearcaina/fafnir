package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode represents different types of application errors
type ErrorCode string

const (
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrForbidden    ErrorCode = "FORBIDDEN"
	ErrInvalidToken ErrorCode = "INVALID_TOKEN"

	ErrInvalidInput ErrorCode = "INVALID_INPUT"

	ErrNotFound ErrorCode = "NOT_FOUND"
	ErrConflict ErrorCode = "CONFLICT"

	ErrInternal ErrorCode = "INTERNAL_ERROR"
	ErrDatabase ErrorCode = "DATABASE_ERROR"
)

// AppError represents a structured application errors
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
	Cause      error     `json:"-"`
}

// Error implements the errors interface
// It returns a string representation of the error, including the code, message, and cause if available
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for errors unwrapping
// This allows the AppError to be used with errors.Is and errors.As
func (e *AppError) Unwrap() error {
	return e.Cause
}

// ToJSON returns the JSON representation for API responses
func (e *AppError) ToJSON() []byte {
	data, _ := json.Marshal(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    e.Code,
			"message": e.Message,
			"details": e.Details,
		},
	})
	return data
}

// New creates a new AppError with the given code and message
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with an AppError, allowing to add additional context
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// Is checks if the given error is some AppError
func Is(err error, target *AppError) bool {
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Code == target.Code
}

// WithDetails adds additional details to an errors
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithHTTPStatus adds the HTTP status (can be overridden if needed)
func (e *AppError) WithHTTPStatus(status int) *AppError {
	e.HTTPStatus = status
	return e
}

// TokenError returns an errors for invalid or expired tokens (will be used if a member tries to access an endpoint with an invalid token)
func TokenError(message string) *AppError {
	return New(ErrInvalidToken, message).WithHTTPStatus(http.StatusUnauthorized)
}

// BadRequestError returns an errors for invalid input (will be used if a member tries to access an endpoint with invalid parameters)
func BadRequestError(message string) *AppError {
	return New(ErrInvalidInput, message).WithHTTPStatus(http.StatusBadRequest)
}

// NotFoundError returns an errors for resources that could not be found (will be used if a member tries to access a resource that does not exist)
func NotFoundError(resource string) *AppError {
	return New(ErrNotFound, resource).WithHTTPStatus(http.StatusNotFound)
}

// ConflictError returns an errors for resource conflicts (will be used if a member tries to create or update a resource that already exists)
func ConflictError(resource string) *AppError {
	return New(ErrConflict, resource).WithHTTPStatus(http.StatusConflict)
}

// UnauthorizedError returns an errors for unauthorized access (will be used if a member tries to access an endpoint or resource without being authenticated)
func UnauthorizedError() *AppError {
	return New(ErrUnauthorized, "Authentication required").WithHTTPStatus(http.StatusUnauthorized)
}

// InternalError returns a generic internal server errors (used for unexpected errors)
func InternalError(message string) *AppError {
	return New(ErrInternal, message).WithHTTPStatus(http.StatusInternalServerError)
}

// DatabaseError wraps a database operation errors (so it can be used to log the original errors)
func DatabaseError(err error) *AppError {
	return Wrap(err, ErrDatabase, "Database operation failed").WithHTTPStatus(http.StatusInternalServerError)
}

// ForbiddenError returns an errors for forbidden access (will be used if a member tries to access an endpoint or resource they are not allowed to)
func ForbiddenError(message string) *AppError {
	return New(ErrForbidden, message).WithHTTPStatus(http.StatusForbidden)
}
