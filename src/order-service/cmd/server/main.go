package main

import (
	"context"
	"fafnir/order-service/internal/api"
	"fafnir/order-service/internal/config"
	"fafnir/order-service/internal/db"
	stockpb "fafnir/shared/pb/stock"
	"fafnir/shared/pkg/logger"
	"fafnir/shared/pkg/nats"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// instantiate custom logger (slog wrapper) for structured logging
	logger := logger.New(nil)

	// instantiate the configuration (environment variables) for the service
	cfg := config.NewConfig()

	// connect to order db
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

	// create stock service client
	stockConn, err := grpc.NewClient(cfg.StockService.URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error(ctx, "Failed to connect to stock service", "error", err)
		os.Exit(1)
	}
	stockClient := stockpb.NewStockServiceClient(stockConn)

	orderHandler := api.NewOrderHandler(db, natsClient, stockClient, logger)

	server := api.NewServer(cfg, logger, orderHandler)

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
		logger.Error(context.Background(), "Order service exited with error", "error", err)
		os.Exit(1)
	}

	logger.Info(context.Background(), "Order service exited cleanly")
}
