package api

import (
	"context"
	"fafnir/api-gateway/graph/generated"
	"fafnir/api-gateway/graph/resolvers"
	"fafnir/api-gateway/internal/config"
	m "fafnir/api-gateway/internal/middleware"
	"log"
	"net/http"

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
	HTTP *http.Server
}

func NewServer() *Server {
	cfg := config.NewConfig()

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
		middleware.Logger,
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
	}
}

func (s *Server) Run() error {
	log.Printf("Starting API gateway on port %s\n", s.HTTP.Addr)
	return s.HTTP.ListenAndServe()
}

func (s *Server) Close(ctx context.Context) error {
	log.Println("Shutting down API gateway gracefully...")

	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		log.Printf("Error during graceful shutdown: %v\n", err)
		return err
	}

	log.Println("API gateway shutdown complete.")
	return nil
}
