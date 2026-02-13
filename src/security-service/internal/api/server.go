package api

import (
	"context"
	"fafnir/security-service/internal/config"
	"fafnir/security-service/internal/db"
	pb "fafnir/shared/pb/security"
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

	// create a db instance
	dbInstance, err := db.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// create a nats client instance
	natsClient, err := nats.New(cfg.NATS.URL, nil) // pass in nil logger for now (TODO: implement for gRPC)
	if err != nil {
		log.Fatal(err)
	}

	// create the security handler
	securityHandler := NewSecurityHandler(dbInstance, natsClient)

	// register subscribe handlers for NATS
	securityHandler.RegisterSubscribeHandlers()

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

	// register the security service with the gRPC server
	pb.RegisterSecurityServiceServer(grpcServer, securityHandler)

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

	log.Printf("Starting gRPC security service on port %s\n", s.config.PORT)

	listener, err := net.Listen("tcp", s.config.PORT)
	if err != nil {
		return err
	}

	return s.grpcServer.Serve(listener)
}

func (s *Server) Close(ctx context.Context) error {
	log.Println("Shutting down security service gracefully...")

	s.grpcServer.GracefulStop()

	log.Println("Security service shutdown complete.")
	return nil
}
