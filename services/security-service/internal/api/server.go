package api

import (
	"context"
	"fafnir/security-service/internal/config"
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

	securityService := NewSecurityService()
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
	log.Printf("Starting auth service on port %s\n", s.HTTP.Addr)
	return s.HTTP.ListenAndServe()
}

func (s *Server) GracefulShutdown(ctx context.Context) error {
	log.Println("Shutting down auth service gracefully...")

	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		log.Printf("Error during graceful shutdown: %v\n", err)
		return err
	}

	log.Println("Auth service shutdown complete.")
	return nil
}
