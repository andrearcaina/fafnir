package main

import (
	"context"
	"fafnir/api-gateway/internal/api"
	"fafnir/api-gateway/internal/config"
	"fafnir/shared/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// instantiate custom logger (slog wrapper) for structured logging
	logger := logger.New(nil)

	// instantiate the configuration (environment variables) for the service
	cfg := config.NewConfig()

	server := api.NewServer(cfg, logger)

	// use errgroup to manage the lifecycle of the server and handle graceful shutdown
	g, ctx := errgroup.WithContext(ctx)

	// start server
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
		logger.Error(context.Background(), "API Gateway exited with error", "error", err)
		os.Exit(1)
	}

	logger.Info(context.Background(), "API Gateway exited cleanly")
}
