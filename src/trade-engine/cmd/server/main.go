package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fafnir/shared/pkg/logger"
	"fafnir/trade-engine/internal/api"
	"fafnir/trade-engine/internal/config"
	"fafnir/trade-engine/internal/engine"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// instantiate custom logger (slog wrapper) for structured logging
	logger := logger.New(nil)

	logger.Info(ctx, "Starting Trade Engine Service")

	cfg := config.New()

	eng, err := engine.NewEngine(cfg, logger)
	if err != nil {
		logger.Error(ctx, "Failed to initialize engine", "error", err)
		os.Exit(1)
	}

	srv := api.NewServer(cfg, logger)
	// use errgroup to manage the lifecycle of the server and handle graceful shutdown
	g, ctx := errgroup.WithContext(ctx)

	// start engine
	g.Go(func() error {
		return eng.Start()
	})

	// start api server
	g.Go(func() error {
		return srv.Run()
	})

	// wait for shutdown signal
	g.Go(func() error {
		<-ctx.Done()

		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Close(closeCtx); err != nil {
			logger.Error(context.Background(), "Server forced to shutdown", "error", err)
		}

		return eng.Stop()
	})

	// wait for everything
	if err := g.Wait(); err != nil {
		logger.Error(context.Background(), "Trade Engine exited with error", "error", err)
		os.Exit(1)
	}

	logger.Info(context.Background(), "Trade Engine exited cleanly")
}
