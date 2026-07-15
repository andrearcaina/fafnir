package resolvers

import (
	"context"
	"errors"

	"fafnir/api-gateway/internal/clients"
)

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	SecurityClient  *clients.SecurityClient
	UserClient      *clients.UserClient
	StockClient     *clients.StockClient
	OrderClient     *clients.OrderClient
	PortfolioClient *clients.PortfolioClient
}

func (r *Resolver) requireOwnedAccounts(ctx context.Context, userID string, accountIDs ...string) error {
	owned, err := r.PortfolioClient.OwnsAccounts(ctx, userID, accountIDs...)
	if err != nil {
		return err
	}
	if !owned {
		return errors.New("portfolio account not found")
	}

	return nil
}
