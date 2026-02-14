package api

import (
	"context"
	pb "fafnir/shared/pb/user"
	"fafnir/shared/pkg/logger"
	"fafnir/user-service/internal/config"
	"fafnir/user-service/internal/db"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer    *grpc.Server
	metricsServer *http.Server
	config        *config.Config
	logger        *logger.Logger
}

func NewServer(cfg *config.Config, db *db.Database, logger *logger.Logger, handler *UserHandler) *Server {
	// register subscribe handlers for NATS
	handler.RegisterSubscribeHandlers()

	// create gRPC server with logging interceptor and prometheus interceptor
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logger.NewGRPCLoggingInterceptor(func(fullMethod string, req interface{}, resp interface{}) map[string]any {
				if fullMethod == "/user.UserService/GetProfileData" {
					if p, ok := resp.(*pb.ProfileDataResponse); ok {
						return map[string]any{
							"user_id": p.GetData().GetUserId(),
						}
					}
				}
				return nil
			}),
			grpc_prometheus.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
		),
	)

	// register the user service with the gRPC server
	pb.RegisterUserServiceServer(grpcServer, handler)

	// register gRPC server metrics
	grpc_prometheus.Register(grpcServer)
	// enable handling of histogram metrics
	grpc_prometheus.EnableHandlingTimeHistogram()

	router := chi.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	metricsServer := &http.Server{
		Addr:    ":9090",
		Handler: router,
	}

	return &Server{
		grpcServer:    grpcServer,
		metricsServer: metricsServer,
		config:        cfg,
		logger:        logger,
	}
}

func (s *Server) RunGRPCServer() error {
	s.logger.Info(context.Background(), "Starting user service", "port", s.config.PORT)

	listener, err := net.Listen("tcp", s.config.PORT)
	if err != nil {
		return err
	}

	return s.grpcServer.Serve(listener)
}

func (s *Server) RunMetricsServer() error {
	s.logger.Info(context.Background(), "Starting metrics server on port 9090")

	if err := s.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Close(ctx context.Context) error {
	s.logger.Info(context.Background(), "Shutting down user service gracefully...")

	s.grpcServer.GracefulStop()

	if err := s.metricsServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("User service shutdown complete.")
	return nil
}
