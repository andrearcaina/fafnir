package api

import (
	"context"
	"fafnir/auth-service/internal/config"
	"fafnir/auth-service/internal/db"
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
	router.Use(middleware.Logger)

	cfg := config.NewConfig()

	// connect to auth db
	dbConn, err := db.NewDBConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v\n", err)
	}

	// create an auth service and handler instance
	authService := NewAuthService(dbConn)
	authHandler := NewAuthHandler(authService)

	// mount the auth handler to the router
	router.Mount("/auth", authHandler.ServeAuthRoutes())

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
