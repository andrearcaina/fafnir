package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fafnir/stock-service/internal/dto"

	"resty.dev/v3"
)

const fmpBaseURL = "https://financialmodelingprep.com/stable"

type FMPProvider struct {
	apiKey string
	client *resty.Client
}

func NewFMP(apiKey string, timeout time.Duration) *FMPProvider {
	return &FMPProvider{
		apiKey: apiKey,
		client: resty.New().
			SetBaseURL(fmpBaseURL).
			SetQueryParam("apikey", apiKey).
			SetTimeout(timeout),
	}
}

func (f *FMPProvider) Close() error {
	return f.client.Close()
}

func (f *FMPProvider) Name() string {
	return "fmp"
}

func (f *FMPProvider) SupportsSymbol(symbol string) bool {
	if f.apiKey == "" {
		return false
	}

	_, ok := fmpSymbols[strings.ToUpper(strings.TrimSpace(symbol))]
	return ok
}

func (f *FMPProvider) GetStockMetadata(ctx context.Context, symbol string) (*dto.StockMetadataResponse, error) {
	var result []dto.StockMetadataResponse

	resp, err := f.client.R().
		SetContext(ctx).
		SetQueryParam("query", symbol).
		SetResult(&result).
		Get("/search-symbol")
	if err != nil {
		return nil, fmt.Errorf("fetch FMP metadata: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("FMP metadata request failed with status %d", resp.StatusCode())
	}

	for index := range result {
		if strings.EqualFold(result[index].Symbol, symbol) {
			if result[index].InstrumentType == "" {
				result[index].InstrumentType = fmpInstrumentType(symbol)
			}

			return &result[index], nil
		}
	}

	return nil, fmt.Errorf("FMP returned no exact metadata match for %s", symbol)
}

func fmpInstrumentType(symbol string) string {
	switch strings.ToUpper(symbol) {
	case "SPY", "SPYG", "VWO":
		return "ETF"
	default:
		return "EQUITY"
	}
}

func (f *FMPProvider) GetStockQuote(ctx context.Context, symbol string) (*dto.StockQuoteResponse, error) {
	type quoteResponse struct {
		dto.StockQuoteResponse
		Timestamp int64 `json:"timestamp"`
	}

	var result []quoteResponse

	resp, err := f.client.R().
		SetContext(ctx).
		SetQueryParam("symbol", symbol).
		SetResult(&result).
		Get("/quote")
	if err != nil {
		return nil, fmt.Errorf("fetch FMP quote: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("FMP quote request failed with status %d", resp.StatusCode())
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("FMP returned no quote for %s", symbol)
	}
	if result[0].LastPrice <= 0 {
		return nil, fmt.Errorf("FMP returned an invalid price for %s", symbol)
	}

	quote := result[0].StockQuoteResponse
	quote.Source = f.Name()
	quote.FetchedAt = time.Now().UTC()
	if result[0].Timestamp > 0 {
		quote.AsOf = time.Unix(result[0].Timestamp, 0).UTC()
	} else {
		quote.AsOf = quote.FetchedAt
	}

	return &quote, nil
}

func (f *FMPProvider) GetStockHistoricalData(ctx context.Context, symbol string, from string, to string) ([]dto.StockHistoricalDataResponse, error) {
	var result []dto.StockHistoricalDataResponse

	resp, err := f.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"symbol": symbol,
			"from":   from,
			"to":     to,
		}).
		SetResult(&result).
		Get("/historical-price-eod/full")
	if err != nil {
		return nil, fmt.Errorf("fetch FMP history: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("FMP history request failed with status %d", resp.StatusCode())
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("FMP returned no historical data for %s", symbol)
	}

	return result, nil
}
