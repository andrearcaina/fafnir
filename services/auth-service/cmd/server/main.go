package main

import (
	"github.com/andrearcaina/den/services/auth-service/internal/api"
	"github.com/andrearcaina/den/services/auth-service/internal/config"
	"log"

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// create a router instance using chi
	router := chi.NewRouter()

	// custom logger middleware (by go chi)
	router.Use(middleware.Logger)

	// create an instance of the auth service and handler
	authService := api.NewAuthService()
	authHandler := api.NewAuthHandler(authService)

	// mount the auth handler to the router
	router.Mount("/auth", authHandler.ServeHTTP())

	// create a config instance for the server
	cfg := config.NewConfig()

	server := &http.Server{
		Addr:    cfg.PORT,
		Handler: router,
	}

	log.Printf("Starting api service on port %v\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
