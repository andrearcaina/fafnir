package provider

import (
	"context"
	"fmt"
	"time"

	"fafnir/stock-service/internal/dto"

	yfclient "github.com/wnjoon/go-yfinance/pkg/client"
	"github.com/wnjoon/go-yfinance/pkg/models"
	yfsearch "github.com/wnjoon/go-yfinance/pkg/search"
	"github.com/wnjoon/go-yfinance/pkg/ticker"
)

type YFProvider struct {
	client   *yfclient.Client
	searcher *yfsearch.Search
}

func NewYahoo(timeout time.Duration) (*YFProvider, error) {
	client, err := yfclient.New(yfclient.WithTimeout(int(timeout.Seconds())))
	if err != nil {
		return nil, fmt.Errorf("create Yahoo Finance client: %w", err)
	}
	searcher, err := yfsearch.New(yfsearch.WithClient(client))
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("create Yahoo Finance search: %w", err)
	}

	return &YFProvider{
		client:   client,
		searcher: searcher,
	}, nil
}

func (y *YFProvider) SearchStocks(ctx context.Context, query string, limit int) ([]dto.StockSearchResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	quotes, err := y.searcher.Quotes(query, limit)
	if err != nil {
		return nil, fmt.Errorf("search Yahoo Finance: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	results := make([]dto.StockSearchResult, 0, len(quotes))
	for _, quote := range quotes {
		if quote.Symbol == "" {
			continue
		}

		name := quote.LongName
		if name == "" {
			name = quote.ShortName
		}

		results = append(results, dto.StockSearchResult{
			Symbol:           quote.Symbol,
			Name:             name,
			Exchange:         quote.Exchange,
			ExchangeFullName: quote.ExchangeDisp,
			InstrumentType:   quote.QuoteType,
		})
	}

	return results, nil
}

func (y *YFProvider) Name() string {
	return "yahoo finance"
}

func (y *YFProvider) Close() {
	y.client.Close()
}

func (y *YFProvider) GetStockMetadata(ctx context.Context, symbol string) (*dto.StockMetadataResponse, error) {
	quote, err := y.quote(ctx, symbol)
	if err != nil {
		return nil, err
	}
	if quote.Symbol == "" || quote.Currency == "" {
		return nil, fmt.Errorf("Yahoo returned incomplete metadata for %s", symbol)
	}

	name := quote.LongName
	if name == "" {
		name = quote.ShortName
	}

	return &dto.StockMetadataResponse{
		Symbol:           quote.Symbol,
		Name:             name,
		Currency:         quote.Currency,
		Exchange:         quote.Exchange,
		ExchangeFullName: quote.ExchangeName,
		InstrumentType:   quote.QuoteType,
	}, nil
}

func (y *YFProvider) GetStockQuote(ctx context.Context, symbol string) (*dto.StockQuoteResponse, error) {
	quote, err := y.quote(ctx, symbol)
	if err != nil {
		return nil, err
	}
	if quote.RegularMarketPrice <= 0 {
		return nil, fmt.Errorf("Yahoo returned an invalid price for %s", symbol)
	}

	now := time.Now().UTC()
	asOf := quote.RegularMarketTime.UTC()
	if asOf.IsZero() || asOf.Unix() <= 0 {
		asOf = now
	}

	return &dto.StockQuoteResponse{
		Symbol:        quote.Symbol,
		LastPrice:     quote.RegularMarketPrice,
		OpenPrice:     quote.RegularMarketOpen,
		PreviousClose: quote.RegularMarketPreviousClose,
		DayLow:        quote.RegularMarketDayLow,
		DayHigh:       quote.RegularMarketDayHigh,
		YearLow:       quote.FiftyTwoWeekLow,
		YearHigh:      quote.FiftyTwoWeekHigh,
		Volume:        quote.RegularMarketVolume,
		MarketCap:     float64(quote.MarketCap),
		Change:        quote.RegularMarketChange,
		ChangePct:     quote.RegularMarketChangePercent,
		Source:        y.Name(),
		AsOf:          asOf,
		FetchedAt:     now,
		MarketState:   quote.MarketState,
		Currency:      quote.Currency,
	}, nil
}

func (y *YFProvider) GetStockHistoricalData(ctx context.Context, symbol string, from string, to string) ([]dto.StockHistoricalDataResponse, error) {
	start, err := time.Parse(time.DateOnly, from)
	if err != nil {
		return nil, fmt.Errorf("parse history start date: %w", err)
	}

	end, err := time.Parse(time.DateOnly, to)
	if err != nil {
		return nil, fmt.Errorf("parse history end date: %w", err)
	}

	stockTicker, err := ticker.New(symbol, ticker.WithClient(y.client))
	if err != nil {
		return nil, fmt.Errorf("create Yahoo ticker: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	bars, err := stockTicker.History(models.HistoryParams{
		Start:      &start,
		End:        &end,
		Interval:   "1d",
		AutoAdjust: false,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch Yahoo history: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := make([]dto.StockHistoricalDataResponse, 0, len(bars))
	for _, bar := range bars {
		change := bar.Close - bar.Open
		var changePct float64
		if bar.Open != 0 {
			changePct = change / bar.Open * 100
		}

		result = append(result, dto.StockHistoricalDataResponse{
			Symbol:     symbol,
			Date:       bar.Date.Format(time.DateOnly),
			OpenPrice:  bar.Open,
			HighPrice:  bar.High,
			LowPrice:   bar.Low,
			ClosePrice: bar.Close,
			Volume:     bar.Volume,
			Change:     change,
			ChangePct:  changePct,
		})
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("Yahoo returned no historical data for %s", symbol)
	}

	return result, nil
}

func (y *YFProvider) quote(ctx context.Context, symbol string) (*models.Quote, error) {
	stockTicker, err := ticker.New(symbol, ticker.WithClient(y.client))
	if err != nil {
		return nil, fmt.Errorf("create Yahoo ticker: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	quote, err := stockTicker.Quote()
	if err != nil {
		return nil, fmt.Errorf("fetch Yahoo quote: %w", err)
	}
	if quote == nil {
		return nil, fmt.Errorf("Yahoo returned no quote for %s", symbol)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return quote, nil
}
