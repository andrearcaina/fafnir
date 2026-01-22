package api

import (
	"context"
	pb "fafnir/shared/pb/stock"
	"fafnir/shared/pkg/redis"
	"fafnir/stock-service/internal/config"
	"fafnir/stock-service/internal/db"
	"fafnir/stock-service/internal/fmp"
	"log"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	config     *config.Config
}

func NewServer() *Server {
	cfg := config.NewConfig()

	// connect to stock db by instantiating a new database connection
	// and passing the config to it
	dbInstance, err := db.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	redisCache, err := redis.New(cfg.Cache)
	if err != nil {
		log.Fatal(err)
	}

	// create a client instance that fetches data from FMP (Financial Modeling Prep) API
	fmpClient, err := fmp.New(cfg.FMP.APIKey)
	if err != nil {
		log.Fatal(err)
	}

	// create a stock service and handler instance
	stockService := NewStockService(dbInstance, redisCache, fmpClient)
	stockHandler := NewStockHandler(stockService)

	// create gRPC server with logging interceptor (interceptors are basically a middleware for gRPC)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	// register the gRPC stock handler with the gRPC server
	pb.RegisterStockServiceServer(grpcServer, stockHandler)

	return &Server{
		grpcServer: grpcServer,
		config:     cfg,
	}
}

func (s *Server) Run() error {
	log.Printf("Starting gRPC stock service on port %s\n", s.config.PORT)

	listener, err := net.Listen("tcp", s.config.PORT)
	if err != nil {
		return err
	}

	return s.grpcServer.Serve(listener)
}

func (s *Server) Close(ctx context.Context) error {
	log.Println("Shutting down stock service gracefully...")

	s.grpcServer.GracefulStop()

	log.Println("Stock service shutdown complete.")
	return nil
}
