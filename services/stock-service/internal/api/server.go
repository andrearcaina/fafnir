package api

import (
	"context"
	"fafnir/stock-service/internal/config"
	"fafnir/stock-service/internal/db"
	"log"
	"net/http"

	"github.com/go-chi/cors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	// connect to stock db by instantiating a new database connection
	// and passing the config to it
	dbInstance, err := db.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// create a stock service and handler instance
	stockService := NewStockService(dbInstance)
	stockHandler := NewStockHandler(stockService)

	// mount the stock handler to the router
	router.Mount("/stock", stockHandler.ServeStockRoutes())

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
	log.Printf("Starting stock service on port %s\n", s.HTTP.Addr)
	return s.HTTP.ListenAndServe()
}

func (s *Server) GracefulShutdown(ctx context.Context) error {
	log.Println("Shutting down stock service gracefully...")

	if s.Database != nil {
		log.Println("Database connection closed.")
		s.Database.Close()
	}

	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		return err
	}

	log.Println("Stock service shutdown complete.")
	return nil
}
