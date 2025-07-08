package main

import (
	"github.com/andrearcaina/den/services/auth-service/internal/handlers"
	"github.com/andrearcaina/den/services/auth-service/internal/service"
	"log"

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// create a mutex
	router := chi.NewRouter()

	// custom logger middleware
	router.Use(middleware.Logger)

	// mount the auth handler with the auth service
	authService := service.NewAuthService()
	authHandler := handlers.NewAuthHandler(authService)
	router.Mount("/auth", authHandler.ServeHTTP())

	server := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	log.Printf("Starting auth service on port %v\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
