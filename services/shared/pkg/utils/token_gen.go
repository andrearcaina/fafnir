package utils

import (
	"crypto/rand"
	"encoding/base64"
	apperrors "fafnir/shared/pkg/errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ParseJWTToken parses the JWT token string and validates it using the provided JWT key, and returns the parsed token if valid
func ParseJWTToken(tokenString string, jwtKey string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.TokenError("Unexpected signing method used").
				WithDetails(fmt.Sprintf("Expected signing method HS256, got %s", token.Header["alg"]))
		}
		return []byte(jwtKey), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, apperrors.TokenError("Invalid token").
			WithDetails("The provided JWT token is invalid or expired")
	}

	return token, nil
}

// GenerateJWTToken creates a new JWT token for the given user ID and signs it with the provided JWT key
func GenerateJWTToken(userID uuid.UUID, jwtKey string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Hour).Unix(), // token valid for 1 hour
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtKey))
}

// GetCSRFTokenFromRequest extracts the CSRF token from the request header
func GetCSRFTokenFromRequest(r *http.Request) string {
	return r.Header.Get("X-CSRF-Token")
}

// GenerateCSRFToken generates a secure CSRF token of the specified length
func GenerateCSRFToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// SetCookie sets a cookie with the specified parameters (can be used to set, update, or delete a cookie depending on maxAge)
func SetCookie(
	w http.ResponseWriter,
	name,
	value string,
	maxAge int,
	httpOnly bool,
	secure bool,
	sameSite http.SameSite,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: httpOnly,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   maxAge,
	})
}
