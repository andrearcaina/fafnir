package main

import (
	"context"
	"fafnir/auth-service/internal/api"
	"fafnir/auth-service/internal/config"
	"fafnir/auth-service/internal/db"
	"fafnir/shared/pkg/logger"
	"fafnir/shared/pkg/nats"
	"fafnir/shared/pkg/validator"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

// main manages the lifecycle of the service, including instantiating the required dependencies for the server
// as well as running the server and gracefully shutting it down with errgroup (built in go pkg for managing goroutines and their errors)
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// instantiate custom logger (slog wrapper) for structured logging
	logger := logger.New(nil)

	// instantiate the configuration (environment variables) for the service
	cfg := config.NewConfig()

	// connect to auth db
	db, err := db.New(cfg, logger)
	if err != nil {
		logger.Error(ctx, "Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// create a nats client instance
	natsClient, err := nats.New(cfg.NATS.URL, logger)
	if err != nil {
		logger.Error(ctx, "Failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	// _, err = natsClient.AddStream("users", []string{"users.>"})
	// if err != nil {
	// 	logger.Error(ctx, "Failed to add NATS stream", "error", err)
	// 	os.Exit(1)
	// }

	// create a custom validator instance for request payload validation
	validator := validator.New()

	// create an auth service and handler instance passing in the db instance, nats client, config, logger and validator
	authService := api.NewAuthService(db, natsClient, cfg.JWT)
	authHandler := api.NewAuthHandler(authService, validator, logger)

	server, err := api.NewServer(cfg, logger, authHandler)
	if err != nil {
		logger.Error(ctx, "Failed to create server", "error", err)
		os.Exit(1)
	}

	// use errgroup to manage the lifecycle of the server and handle graceful shutdown
	g, ctx := errgroup.WithContext(ctx)

	// start HTTP server
	g.Go(func() error {
		return server.Run()
	})

	// wait for shutdown signal
	g.Go(func() error {
		<-ctx.Done()

		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return server.Close(closeCtx)
	})

	// wait for everything
	if err := g.Wait(); err != nil {
		logger.Error(context.Background(), "Auth service exited with error", "error", err)
		os.Exit(1)
	}

	logger.Info(context.Background(), "Auth service exited cleanly")
}
