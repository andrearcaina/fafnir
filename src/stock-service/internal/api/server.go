package api

import (
	"context"
	pb "fafnir/shared/pb/stock"
	"fafnir/shared/pkg/logger"
	"fafnir/stock-service/internal/config"
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

func NewServer(cfg *config.Config, logger *logger.Logger, handler *StockHandler) *Server {
	// create gRPC server with logging interceptor and prometheus interceptor
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logger.NewGRPCLoggingInterceptor(nil),
			grpc_prometheus.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
		),
	)

	// register the gRPC stock handler with the gRPC server
	pb.RegisterStockServiceServer(grpcServer, handler)

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
	s.logger.Info(context.Background(), "Starting stock service", "port", s.config.PORT)

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
	s.logger.Info(context.Background(), "Shutting down stock service gracefully...")

	s.grpcServer.GracefulStop()

	if err := s.metricsServer.Shutdown(ctx); err != nil {
		return err
	}

	s.logger.Info(context.Background(), "Stock service shutdown complete.")
	return nil
}
