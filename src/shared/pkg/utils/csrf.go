package utils

import (
	"crypto/subtle"
	apperrors "fafnir/shared/pkg/errors"
	"net/http"
)

func ValidateCSRFToken(r *http.Request) error {
	if isSafeMethod(r.Method) {
		return nil
	}

	csrfCookie, err := r.Cookie("csrf_token")
	if err != nil {
		return apperrors.UnauthorizedError().
			WithDetails("CSRF token not found in cookies")
	}

	requestToken := GetCSRFTokenFromRequest(r)
	if requestToken == "" || subtle.ConstantTimeCompare(
		[]byte(csrfCookie.Value),
		[]byte(requestToken),
	) != 1 {
		return apperrors.UnauthorizedError().
			WithDetails("Invalid CSRF token")
	}

	return nil
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
