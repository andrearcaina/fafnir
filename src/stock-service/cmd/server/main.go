package main

import (
	"context"
	"fafnir/shared/pkg/logger"
	"fafnir/shared/pkg/redis"
	"fafnir/stock-service/internal/api"
	"fafnir/stock-service/internal/config"
	"fafnir/stock-service/internal/db"
	"fafnir/stock-service/internal/fmp"
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

	// connect to stock db
	dbInstance, err := db.New(cfg, logger)
	if err != nil {
		logger.Error(ctx, "Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// create redis cache
	redisCache, err := redis.New(cfg.Cache, logger)
	if err != nil {
		logger.Error(ctx, "Failed to initialize redis", "error", err)
		os.Exit(1)
	}

	// create FMP client
	fmpClient, err := fmp.New(cfg.FMP.APIKey)
	if err != nil {
		logger.Error(ctx, "Failed to initialize FMP client", "error", err)
		os.Exit(1)
	}

	stockService := api.NewStockService(dbInstance, redisCache, fmpClient)
	stockHandler := api.NewStockHandler(stockService, logger)

	server := api.NewServer(cfg, logger, stockHandler)

	// use errgroup to manage the lifecycle of the server and handle graceful shutdown
	g, ctx := errgroup.WithContext(ctx)

	// start GRPC server
	g.Go(func() error {
		return server.RunGRPCServer()
	})

	// start metrics server
	g.Go(func() error {
		return server.RunMetricsServer()
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
		logger.Error(context.Background(), "Stock service exited with error", "error", err)
		os.Exit(1)
	}

	logger.Info(context.Background(), "Stock service exited cleanly")
}
