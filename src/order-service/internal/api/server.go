package api

import (
	"context"
	"fafnir/order-service/internal/config"
	"fafnir/order-service/internal/db"
	pb "fafnir/shared/pb/order"
	stockpb "fafnir/shared/pb/stock"
	"fafnir/shared/pkg/nats"

	"log"
	"net"
	"net/http"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	grpcServer *grpc.Server
	config     *config.Config
}

func NewServer() *Server {
	cfg := config.NewConfig()

	dbInstance, err := db.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// create a nats client instance
	natsClient, err := nats.New(cfg.NATS.URL, nil) // pass in nil logger for now (TODO: implement for gRPC)
	if err != nil {
		log.Fatal(err)
	}

	// create stock service client
	stockConn, err := grpc.NewClient(cfg.StockService.URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	stockClient := stockpb.NewStockServiceClient(stockConn)

	orderHandler := NewOrderHandler(dbInstance, natsClient, stockClient)
	orderHandler.RegisterSubscribeHandlers()

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

	// register the gRPC order handler with the gRPC server
	pb.RegisterOrderServiceServer(grpcServer, orderHandler)

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

	log.Printf("Starting gRPC order service on port %s\n", s.config.PORT)

	listener, err := net.Listen("tcp", s.config.PORT)
	if err != nil {
		return err
	}

	return s.grpcServer.Serve(listener)
}

func (s *Server) Close(ctx context.Context) error {
	log.Println("Shutting down order service gracefully...")

	s.grpcServer.GracefulStop()

	log.Println("Order service shutdown complete.")
	return nil
}
