package api

import (
	"context"
	"fafnir/portfolio-service/internal/config"
	"fafnir/portfolio-service/internal/db"
	portfoliopb "fafnir/shared/pb/portfolio"
	"fafnir/shared/pkg/nats"
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

	dbInstance, err := db.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	natsClient, err := nats.New(cfg.NATS.URL)
	if err != nil {
		log.Fatal(err)
	}

	handler := NewPortfolioHandler(dbInstance, natsClient)
	handler.RegisterSubscribeHandlers()

	// create gRPC server with logging interception and prometheus interception
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor,
			grpc_prometheus.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
		),
	)

	// register the gRPC handler
	portfoliopb.RegisterPortfolioServiceServer(grpcServer, handler)

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

	log.Printf("Starting gRPC portfolio service on port %s\n", s.config.PORT)

	listener, err := net.Listen("tcp", s.config.PORT)
	if err != nil {
		return err
	}

	return s.grpcServer.Serve(listener)
}

func (s *Server) Close(ctx context.Context) error {
	log.Println("Shutting down portfolio service gracefully...")
	s.grpcServer.GracefulStop()
	log.Println("Portfolio service shutdown complete.")
	return nil
}
