package api

import (
	"context"
	"fafnir/api-gateway/graph/generated"
	"fafnir/api-gateway/graph/resolvers"
	"fafnir/api-gateway/internal/config"
	m "fafnir/api-gateway/internal/middleware"
	"net/http"

	"fafnir/shared/pkg/logger"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vektah/gqlparser/v2/ast"
)

type Server struct {
	HTTP   *http.Server
	config *config.Config
	logger *logger.Logger
}

func NewServer(cfg *config.Config, logger *logger.Logger) *Server {
	srv := handler.New(generated.NewExecutableSchema(
		generated.Config{
			Resolvers: &resolvers.Resolver{
				SecurityClient:  cfg.CLIENTS.SecurityClient,
				UserClient:      cfg.CLIENTS.UserClient,
				StockClient:     cfg.CLIENTS.StockClient,
				OrderClient:     cfg.CLIENTS.OrderClient,
				PortfolioClient: cfg.CLIENTS.PortfolioClient,
			},
		},
	))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	router := chi.NewRouter()

	// global middlewares (chi default middlewares)
	router.Use(
		// middleware.Logger,
		logger.RequestLogger, // TODO: make it showcase actual graphql "endpoints"
		middleware.Recoverer,
	)

	// reverse proxy for auth service
	router.Mount("/auth/", m.ReverseProxy(cfg.PROXY.TargetURL))

	// graphql endpoints for core services (/ is the playground, which is like a UI to test queries)
	// while /graphql is the actual endpoint to send queries and mutations
	router.Handle("/", playground.Handler("GraphQL playground", "/graphql"))
	// prometheus endpoint for monitoring
	router.Handle("/metrics", promhttp.Handler())
	router.With(
		m.ValidateAuth(cfg.ENV.JWT, true),
	).Handle("/graphql", srv)

	return &Server{
		HTTP: &http.Server{
			Addr:    cfg.PORT,
			Handler: router,
		},
		config: cfg,
		logger: logger,
	}
}

func (s *Server) Run() error {
	s.logger.Info(context.Background(), "Starting auth service", "port", s.HTTP.Addr)

	if err := s.HTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Close(ctx context.Context) error {
	s.logger.Info(context.Background(), "Shutting down API gateway gracefully...")

	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		s.logger.Error(context.Background(), "Error during graceful shutdown", "error", err)
		return err
	}

	s.logger.Info(context.Background(), "API gateway shutdown complete.")
	return nil
}
