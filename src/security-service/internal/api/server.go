package api

import (
	"context"
	"fafnir/security-service/internal/config"
	"fafnir/security-service/internal/db"
	pb "fafnir/shared/pb/security"
	"fafnir/shared/pkg/nats"
	"log"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	handler    *SecurityHandler
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

	// create a nats client instanc
	natsClient, err := nats.New(cfg.NATS.URL)
	if err != nil {
		log.Fatal(err)
	}

	// create the security handler
	securityHandler := NewSecurityHandler(dbInstance, natsClient)

	// register subscribe handlers for NATS
	securityHandler.RegisterSubscribeHandlers()

	// create gRPC server with logging interceptor (interceptors are basically a middleware for gRPC)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	// register the security service with the gRPC server
	pb.RegisterSecurityServiceServer(grpcServer, securityHandler)

	return &Server{
		handler:    securityHandler,
		grpcServer: grpcServer,
		config:     cfg,
	}
}

func (s *Server) Run() error {
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
