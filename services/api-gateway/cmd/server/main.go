package main

import (
	"fafnir/api-gateway/internal/clients"
	"fafnir/api-gateway/internal/config"
	"fafnir/api-gateway/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func main() {
	authClient := clients.NewAuthClient("http://fafnir-auth-service-1:8081/")

	handlerConfig := &handlers.HandlerConfig{
		AuthServiceClient: authClient,
		// add more clients for services later on
	}

	graphQLHandler, err := handlers.NewGraphQLHandler(handlerConfig)

	if err != nil {
		log.Fatalf("Failed to create GraphQL handler: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Handle("/graphql", graphQLHandler)

	conf := config.NewConfig()

	server := &http.Server{
		Addr:    conf.PORT,
		Handler: r,
	}

	log.Printf("Starting API Gateway on port %v\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
