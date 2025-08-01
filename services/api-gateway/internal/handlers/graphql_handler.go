package handlers

import (
	"fafnir/api-gateway/internal/clients"
	"github.com/graphql-go/handler"
	"net/http"
)

type HandlerConfig struct {
	AuthServiceClient *clients.AuthClient
}

func NewGraphQLHandler(handlerConfig *HandlerConfig) (http.Handler, error) {
	authSchema, err := createGraphQLAuthSchema(handlerConfig.AuthServiceClient)
	if err != nil {
		return nil, err
	}

	// add more schemas for other services later on

	h := handler.New(&handler.Config{
		Schema:     &authSchema,
		Pretty:     true,
		GraphiQL:   true,
		Playground: true,
	})

	return h, nil
}
