package main

import (
	"net/http"

	handler "github.com/andrearcaina/den/internal/handler/http/auth"
	"github.com/andrearcaina/den/internal/middleware"
	service "github.com/andrearcaina/den/internal/service/auth"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("Failed to create logger: " + err.Error())
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			panic("Failed to sync logger: " + err.Error())
		} else {
			logger.Info("Logger synced successfully")
		}
	}(logger)

	// create a mutex
	r := chi.NewRouter()

	// custom logger middleware
	r.Use(middleware.Logger(logger))

	// mount the auth handler with the auth service
	authService := service.NewAuthService()
	authHandler := handler.NewAuthHandler(authService)
	r.Mount("/auth", authHandler.ServeHTTP())

	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
