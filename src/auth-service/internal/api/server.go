package api

import (
	"context"
	"fafnir/auth-service/internal/config"
	"fafnir/shared/pkg/logger"
	"net/http"

	"github.com/go-chi/cors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	HTTP   *http.Server
	logger *logger.Logger
}

func NewServer(cfg *config.Config, logger *logger.Logger, authHandler *Handler) (*Server, error) {
	router := chi.NewRouter()

	// set up CORS options
	corsOptions := cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5000", "http://localhost:9090"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	// custom logger middleware (by go chi)
	router.Use(
		corsOptions,
		// middleware.Logger,
		logger.RequestLogger,
		middleware.Recoverer,
	)

	// mount the auth handler to the router at /auth path
	router.Mount("/auth", authHandler.ServeAuthRoutes())
	router.Handle("/metrics", promhttp.Handler())

	return &Server{
		HTTP: &http.Server{
			Addr:    cfg.PORT,
			Handler: router,
		},
		logger: logger,
	}, nil
}

func (s *Server) Run() error {
	s.logger.Info(context.Background(), "Starting auth service", "port", s.HTTP.Addr)

	if err := s.HTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Close(ctx context.Context) error {
	s.logger.Info(ctx, "Shutting down auth service gracefully...")

	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		return err
	}

	s.logger.Info(ctx, "Auth service shutdown complete.")
	return nil
}
