package api

import (
	"context"
	"fafnir/auth-service/internal/config"
	"fafnir/auth-service/internal/db"
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
}

func NewServer() *Server {
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
		middleware.Logger,
		middleware.Recoverer,
	)

	cfg := config.NewConfig()

	// connect to auth db by instantiating a new database connection
	// and passing the config to it
	dbInstance, err := db.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// create a nats client instance
	natsClient, err := nats.New(cfg.NATS.URL)
	if err != nil {
		log.Fatal(err)
	}
	// _, err = natsClient.AddStream("users", []string{"users.>"})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// create a custom validator instance for request payload validation
	validator := validator.New()

	// create an auth service and handler instance passing in the db instance, nats client, config and validator
	authService := NewAuthService(dbInstance, natsClient, cfg.JWT)
	authHandler := NewAuthHandler(authService, validator)

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
	}
}

func (s *Server) Run() error {
	log.Printf("Starting auth service on port %s\n", s.HTTP.Addr)
	return s.HTTP.ListenAndServe()
}

func (s *Server) Close(ctx context.Context) error {
	log.Println("Shutting down auth service gracefully...")

	if s.Database != nil {
		log.Println("Database connection closed.")
		s.Database.Close()
	}

	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		return err
	}

	log.Println("Auth service shutdown complete.")
	return nil
}
