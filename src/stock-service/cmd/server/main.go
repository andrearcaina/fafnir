package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fafnir/shared/pkg/logger"
	"fafnir/shared/pkg/redis"
	"fafnir/stock-service/internal/api"
	"fafnir/stock-service/internal/config"
	"fafnir/stock-service/internal/db"
	"fafnir/stock-service/internal/provider"

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
	db, err := db.New(cfg, logger)
	if err != nil {
		logger.Error(ctx, "Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// create redis cache
	redisCache, err := redis.New(cfg.Cache, logger)
	if err != nil {
		logger.Error(ctx, "Failed to initialize redis", "error", err)
		os.Exit(1)
	}
	defer redisCache.Close()

	fmpProvider := provider.NewFMP(cfg.FMP.APIKey, cfg.FMP.Timeout)
	defer func() {
		if err := fmpProvider.Close(); err != nil {
			logger.Error(context.Background(), "Failed to close FMP provider", "error", err)
		}
	}()
	yahooProvider, err := provider.NewYahoo(cfg.YahooTimeout)
	if err != nil {
		logger.Error(ctx, "Failed to initialize Yahoo Finance provider", "error", err)
		os.Exit(1)
	}
	defer yahooProvider.Close()

	marketData := provider.NewChain(
		yahooProvider,
		fmpProvider,
	)

	stockService := api.NewStockService(db, redisCache, marketData, yahooProvider, cfg.QuoteTTL)
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
