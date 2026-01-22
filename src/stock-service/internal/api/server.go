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
	"net/http"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// create gRPC server with logging interceptor and prometheus interceptor
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor,
			grpc_prometheus.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
		),
	)

	// register the gRPC stock handler with the gRPC server
	pb.RegisterStockServiceServer(grpcServer, stockHandler)

	// register gRPC server metrics
	grpc_prometheus.Register(grpcServer)
	// enable handling of histogram metrics
	grpc_prometheus.EnableHandlingTimeHistogram()

	return &Server{
		grpcServer: grpcServer,
		config:     cfg,
	}
}

func (s *Server) Run() error {
	// start metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Starting metrics server on port :9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

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
