package api

import (
	"context"
	pb "fafnir/shared/pb/user"
	"fafnir/shared/pkg/nats"
	"fafnir/user-service/internal/config"
	"fafnir/user-service/internal/db"
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

	// create a db instance
	dbInstance, err := db.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// create a nats client instance
	natsClient, err := nats.New(cfg.NATS.URL)
	if err != nil {
		log.Fatal(err)
	}

	// create the user handler
	userHandler := NewUserHandler(dbInstance, natsClient)

	// register subscribe handlers for NATS
	userHandler.RegisterSubscribeHandlers()

	// create gRPC server with logging interceptor (interceptors are basically a middleware for gRPC)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	// register the user service with the gRPC server
	pb.RegisterUserServiceServer(grpcServer, userHandler)

	return &Server{
		grpcServer: grpcServer,
		config:     cfg,
	}
}

func (s *Server) Run() error {
	log.Printf("Starting gRPC user service on port %s\n", s.config.PORT)

	listener, err := net.Listen("tcp", s.config.PORT)
	if err != nil {
		return err
	}

	return s.grpcServer.Serve(listener)
}

func (s *Server) Close(ctx context.Context) error {
	log.Println("Shutting down user service gracefully...")

	s.grpcServer.GracefulStop()

	log.Println("User service shutdown complete.")
	return nil
}
