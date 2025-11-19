package api

import (
	apperrors "fafnir/shared/pkg/errors"
	"regexp"
	"strings"
)

var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type AuthRequest interface {
	GetEmail() string
	GetPassword() string
}

func ValidateAuthRequest(req AuthRequest) error {
	email := req.GetEmail()
	password := req.GetPassword()

	// Check if both are missing
	if email == "" && password == "" {
		return apperrors.BadRequestError("Email and password are required")
	}

	// Validate email
	if err := validateEmail(email); err != nil {
		return err
	}

	// Validate password
	if err := validatePassword(password); err != nil {
		return err
	}

	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return apperrors.BadRequestError("Email is required")
	}

	if !EmailRegex.MatchString(email) {
		return apperrors.BadRequestError("Email must be a valid email address").
			WithDetails("The provided email has an invalid format")
	}

	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return apperrors.BadRequestError("Password is required")
	}

	if len(password) < 8 {
		return apperrors.BadRequestError("Password must be at least 8 characters long").
			WithDetails("The provided password has an invalid format")
	}

	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasDigit := strings.ContainsAny(password, "0123456789")

	if !hasUpper || !hasLower || !hasDigit {
		return apperrors.BadRequestError("Password does not meet strength requirements").
			WithDetails("Password must contain at least one uppercase letter, one lowercase letter, and one digit")
	}

	return nil
}
