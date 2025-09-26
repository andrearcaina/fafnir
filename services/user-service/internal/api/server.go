package api

import (
	"context"
	"fafnir/shared/pb/user"
	"fafnir/user-service/internal/config"
	"fafnir/user-service/internal/db"
	"log"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	handler    *UserHandler
	grpcServer *grpc.Server
	config     *config.Config
}

func NewServer() *Server {
	cfg := config.NewConfig()

	// just to make sure the database connection is established (will assign a var later once we have a service)
	dbInstance, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// create the user handler
	userHandler := NewUserHandler(dbInstance)

	// create gRPC server with logging interceptor (interceptors are basically a middleware for gRPC)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	// register the user service with the gRPC server
	pb.RegisterUserServiceServer(grpcServer, userHandler)

	return &Server{
		handler:    userHandler,
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

func (s *Server) GracefulShutdown(ctx context.Context) error {
	log.Println("Shutting down user service gracefully...")

	s.grpcServer.GracefulStop()

	log.Println("User service shutdown complete.")
	return nil
}
