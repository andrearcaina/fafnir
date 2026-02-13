package api

import (
	"context"
	"fafnir/auth-service/internal/config"
	"fafnir/auth-service/internal/db"
	"fafnir/shared/pkg/logger"
	"fafnir/shared/pkg/nats"
	"fafnir/shared/pkg/validator"
	"log"
	"net/http"

	"github.com/go-chi/cors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	HTTP     *http.Server
	Database *db.Database
	Logger   *logger.Logger
}

func NewServer(logger *logger.Logger) *Server {
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

	cfg := config.NewConfig()

	// connect to auth db by instantiating a new database connection
	// and passing the config to it
	dbInstance, err := db.New(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}

	// create a nats client instance
	natsClient, err := nats.New(cfg.NATS.URL, logger)
	if err != nil {
		log.Fatal(err)
	}
	// _, err = natsClient.AddStream("users", []string{"users.>"})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// create a custom validator instance for request payload validation
	validator := validator.New()

	// create an auth service and handler instance passing in the db instance, nats client, config, logger and validator
	authService := NewAuthService(dbInstance, natsClient, cfg.JWT)
	authHandler := NewAuthHandler(authService, validator, logger)

	// mount the auth handler to the router at /auth path
	router.Mount("/auth", authHandler.ServeAuthRoutes())
	router.Handle("/metrics", promhttp.Handler())

	// create a config instance for the server
	return &Server{
		HTTP: &http.Server{
			Addr:    cfg.PORT,
			Handler: router,
		},
		Database: dbInstance,
		Logger:   logger,
	}
}

func (s *Server) Run() error {
	s.Logger.Info(context.Background(), "Starting auth service", "port", s.HTTP.Addr)
	return s.HTTP.ListenAndServe()
}

func (s *Server) Close(ctx context.Context) error {
	s.Logger.Info(ctx, "Shutting down auth service gracefully...")

	if s.Database != nil {
		s.Logger.Info(ctx, "Database connection closed.")
		s.Database.Close()
	}

	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		return err
	}

	s.Logger.Info(ctx, "Auth service shutdown complete.")
	return nil
}
