package resolvers

import "fafnir/api-gateway/internal/clients"

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	SecurityClient *clients.SecurityClient
	UserClient     *clients.UserClient
	StockClient    *clients.StockClient
	OrderClient    *clients.OrderClient
}
