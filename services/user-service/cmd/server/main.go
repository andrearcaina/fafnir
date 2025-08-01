package main

import (
	"fafnir/user-service/internal/api"
	"fafnir/user-service/internal/config"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	userService := api.NewUserService()
	userHandler := api.NewUserHandler(userService)

	router.Mount("/user", userHandler.ServeUserRoutes())

	cfg := config.NewConfig()

	server := &http.Server{
		Addr:    cfg.PORT,
		Handler: router,
	}

	log.Printf("Starting user service on port %v\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
