package api

import (
	"context"
	"fafnir/security-service/internal/config"
	"fafnir/security-service/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

type Server struct {
	HTTP *http.Server
}

func NewServer() *Server {
	router := chi.NewRouter()

	// custom logger middleware (by go chi)
	router.Use(
		middleware.Logger,
		middleware.Recoverer,
	)

	cfg := config.NewConfig()

	// just to make sure the database connection is established (will assign a var later once we have a service)
	_, err := db.NewDBConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// this will be handled differently using gRPC in the future
	// for now, we will use a simple HTTP handler for REST API calls
	securityService := NewSecurityService( /* pass in connections conn later */ )
	securityHandler := NewSecurityHandler(securityService)

	// mount the auth handler to the router
	router.Mount("/security", securityHandler.ServeSecurityRoutes())

	// create a config instance for the server
	return &Server{
		HTTP: &http.Server{
			Addr:    cfg.PORT,
			Handler: router,
		},
	}
}

func (s *Server) Run() error {
	log.Printf("Starting security service on port %s\n", s.HTTP.Addr)
	return s.HTTP.ListenAndServe()
}

func (s *Server) GracefulShutdown(ctx context.Context) error {
	log.Println("Shutting down security service gracefully...")

	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		log.Printf("Error during graceful shutdown: %v\n", err)
		return err
	}

	log.Println("Security service shutdown complete.")
	return nil
}
