package handlers

import (
	"fafnir/api-gateway/internal/clients"
	"fafnir/api-gateway/internal/resolvers"
	"fafnir/api-gateway/internal/schema"
	"fmt"
	"github.com/graphql-go/graphql"
)

func createGraphQLAuthSchema(authServiceClient *clients.AuthClient) (graphql.Schema, error) {
	rootQuery := schema.QueryType

	healthField, ok := rootQuery.Fields()["health"]
	if !ok {
		return graphql.Schema{}, fmt.Errorf("health field not found in QueryType, check schema definition")
	}
	healthField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
		return "API Gateway is running", nil
	}

	mutationResolver := resolvers.NewMutationResolver(authServiceClient)

	rootMutation := schema.MutationType

	loginField, ok := rootMutation.Fields()["login"]
	if !ok {
		return graphql.Schema{}, fmt.Errorf("login field not found in MutationType, check schema definition")
	}
	loginField.Resolve = mutationResolver.LoginResolver

	schemaConfig := graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	}

	graphqlSchema, err := schema.NewAuthSchema(schemaConfig)
	if err != nil {
		return graphql.Schema{}, err
	}

	return graphqlSchema, nil
}
