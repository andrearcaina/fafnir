package main

import (
	"github.com/andrearcaina/den/services/auth-service/internal/handlers"
	"github.com/andrearcaina/den/services/auth-service/internal/service"

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// create a mutex
	r := chi.NewRouter()

	// custom logger middleware
	r.Use(middleware.Logger)

	// mount the auth handler with the auth service
	authService := service.NewAuthService()
	authHandler := handlers.NewAuthHandler(authService)
	r.Mount("/auth", authHandler.ServeHTTP())

	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
