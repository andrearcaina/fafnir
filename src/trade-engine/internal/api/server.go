package api

import (
	"context"
	"log"
	"net/http"

	"fafnir/trade-engine/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	HTTP *http.Server
}

func NewServer(cfg *config.Config) *Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health")) // health check endpoint given by chi middleware

	return &Server{
		HTTP: &http.Server{
			Addr:    cfg.PORT,
			Handler: r,
		},
	}
}

func (s *Server) Run() error {
	log.Printf("Starting trade-engine on port %s\n", s.HTTP.Addr)
	return s.HTTP.ListenAndServe()
}

func (s *Server) Close(ctx context.Context) error {
	log.Println("Shutting down trade-engine gracefully...")

	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		return err
	}

	log.Println("Trade-engine shutdown complete.")
	return nil
}
