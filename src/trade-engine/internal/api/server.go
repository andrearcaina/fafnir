package api

import (
	"context"
	"net/http"

	"fafnir/shared/pkg/logger"
	"fafnir/trade-engine/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	HTTP   *http.Server
	Logger *logger.Logger
}

func NewServer(cfg *config.Config, logger *logger.Logger) *Server {
	r := chi.NewRouter()

	r.Use(
		// middleware.Logger,
		logger.RequestLogger,
		middleware.Recoverer,
		middleware.Heartbeat("/health"), // health check endpoint given by chi middleware
	)

	return &Server{
		HTTP: &http.Server{
			Addr:    cfg.PORT,
			Handler: r,
		},
		Logger: logger,
	}
}

func (s *Server) Run() error {
	s.Logger.Info(context.Background(), "Starting trade-engine", "port", s.HTTP.Addr)

	if err := s.HTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Close(ctx context.Context) error {
	s.Logger.Info(context.Background(), "Shutting down trade-engine gracefully...")

	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		return err
	}

	s.Logger.Info(context.Background(), "Trade-engine shutdown complete.")
	return nil
}
