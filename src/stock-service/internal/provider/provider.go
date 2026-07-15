package provider

import (
	"context"
	"errors"
	"fmt"

	"fafnir/stock-service/internal/dto"
)

type MarketData interface {
	Name() string
	GetStockMetadata(context.Context, string) (*dto.StockMetadataResponse, error)
	GetStockQuote(context.Context, string) (*dto.StockQuoteResponse, error)
	GetStockHistoricalData(context.Context, string, string, string) ([]dto.StockHistoricalDataResponse, error)
}

type SymbolSearcher interface {
	SearchStocks(context.Context, string, int) ([]dto.StockSearchResult, error)
}

type SymbolConstrained interface {
	SupportsSymbol(string) bool
}

type Chain struct {
	providers []MarketData
}

var errNoProviders = errors.New("no market data providers configured")

func NewChain(providers ...MarketData) *Chain {
	return &Chain{
		providers: providers,
	}
}

func (c *Chain) Name() string {
	return "provider-chain"
}

func (c *Chain) GetStockMetadata(ctx context.Context, symbol string) (*dto.StockMetadataResponse, error) {
	return firstResult(ctx, c.providers, symbol, func(dataProvider MarketData) (*dto.StockMetadataResponse, error) {
		return dataProvider.GetStockMetadata(ctx, symbol)
	})
}

func (c *Chain) GetStockQuote(ctx context.Context, symbol string) (*dto.StockQuoteResponse, error) {
	return firstResult(ctx, c.providers, symbol, func(dataProvider MarketData) (*dto.StockQuoteResponse, error) {
		return dataProvider.GetStockQuote(ctx, symbol)
	})
}

func (c *Chain) GetStockHistoricalData(ctx context.Context, symbol string, from string, to string) ([]dto.StockHistoricalDataResponse, error) {
	return firstResult(ctx, c.providers, symbol, func(dataProvider MarketData) ([]dto.StockHistoricalDataResponse, error) {
		return dataProvider.GetStockHistoricalData(ctx, symbol, from, to)
	})
}

func firstResult[T any](ctx context.Context, providers []MarketData, symbol string, fetch func(MarketData) (T, error)) (T, error) {
	var zero T
	if len(providers) == 0 {
		return zero, errNoProviders
	}

	var errs []error

	for _, dataProvider := range providers {
		if !supportsSymbol(dataProvider, symbol) {
			continue
		}

		result, err := fetch(dataProvider)
		if err == nil {
			return result, nil
		}

		if ctx.Err() != nil {
			return zero, ctx.Err()
		}

		errs = append(errs, fmt.Errorf("%s: %w", dataProvider.Name(), err))
	}

	return zero, providerErrors(symbol, errs)
}

func supportsSymbol(dataProvider MarketData, symbol string) bool {
	constrained, ok := dataProvider.(SymbolConstrained)
	return !ok || constrained.SupportsSymbol(symbol)
}

func providerErrors(symbol string, errs []error) error {
	if len(errs) == 0 {
		return fmt.Errorf("no provider supports %s", symbol)
	}

	return errors.Join(errs...)
}
