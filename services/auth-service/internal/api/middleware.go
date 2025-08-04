package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
)

type contextKey string

const userIDContextKey contextKey = "auth/userId"

func CheckAuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				http.Error(w, "Unauthorized: no cookie", http.StatusUnauthorized)
				return
			}

			token, err := parseJWTToken(cookie.Value, jwtSecret)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Unauthorized: no claims", http.StatusUnauthorized)
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok {
				http.Error(w, "Unauthorized: no subject", http.StatusUnauthorized)
				return
			}

			// Add userID to context
			ctx := context.WithValue(r.Context(), userIDContextKey, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func parseJWTToken(tokenString, jwtSecret string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

func GetUserIdFromContext(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value(userIDContextKey)
	userIDStr, ok := val.(string)
	if !ok {
		return uuid.Nil, errors.New("user Id not found in context")
	}

	parsed, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, err
	}

	return parsed, nil
}
