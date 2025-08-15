package api

import (
	"context"
	apperrors "fafnir/shared/pkg/errors"
	"fafnir/shared/pkg/utils"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

func CheckAuth(authService *Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtCookie, err := r.Cookie("auth_token")
			if err != nil {
				err := apperrors.UnauthorizedError().
					WithDetails("Authentication token not found in cookies")
				utils.HandleError(w, err)
				return
			}

			csrfCookie, err := r.Cookie("csrf_token")
			if err != nil {
				err := apperrors.UnauthorizedError().
					WithDetails("CSRF token not found in cookies")
				utils.HandleError(w, err)
				return
			}

			if csrfCookie.Value != utils.GetCSRFTokenFromRequest(r) {
				err := apperrors.UnauthorizedError().
					WithDetails("Invalid CSRF token")
				utils.HandleError(w, err)
				return
			}

			token, err := utils.ParseJWTToken(jwtCookie.Value, authService.jwtKey)
			if err != nil {
				err := apperrors.UnauthorizedError().
					WithDetails("Invalid authentication token")
				utils.HandleError(w, err)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				err := apperrors.UnauthorizedError().
					WithDetails("Invalid token claims")
				utils.HandleError(w, err)
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok {
				err := apperrors.UnauthorizedError().
					WithDetails("Token subject is missing or invalid")
				utils.HandleError(w, err)
				return
			}

			// Add userID to context
			ctx := context.WithValue(r.Context(), contextKey("auth/userIdContextKey"), sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIdFromContext(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value(contextKey("auth/userIdContextKey"))
	userIDStr, ok := val.(string)
	if !ok {
		return uuid.Nil, apperrors.UnauthorizedError().
			WithDetails("User ID not found in context")
	}

	parsed, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, apperrors.ValidationError("Invalid user ID format").
			WithDetails("User ID must be a valid UUID")
	}

	return parsed, nil
}
