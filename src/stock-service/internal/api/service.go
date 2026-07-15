package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"fafnir/shared/pkg/errors"
	"fafnir/shared/pkg/redis"
	"fafnir/stock-service/internal/db"
	"fafnir/stock-service/internal/db/generated"
	"fafnir/stock-service/internal/dto"
	"fafnir/stock-service/internal/provider"
	"fafnir/stock-service/internal/utils"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

type Service struct {
	db           *db.Database
	redis        *redis.Cache
	marketData   provider.MarketData
	symbolSearch provider.SymbolSearcher
	quoteTTL     time.Duration
	requestGroup singleflight.Group
}

func NewStockService(database *db.Database, redis *redis.Cache, marketData provider.MarketData, symbolSearch provider.SymbolSearcher, quoteTTL time.Duration) *Service {
	return &Service{
		db:           database,
		redis:        redis,
		marketData:   marketData,
		symbolSearch: symbolSearch,
		quoteTTL:     quoteTTL,
	}
}

func (s *Service) SearchStocks(ctx context.Context, query string, limit int) ([]dto.StockSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.BadRequestError("Invalid search query").
			WithDetails("The search query is empty")
	}
	if limit <= 0 || limit > 20 {
		limit = 8
	}

	results, err := s.symbolSearch.SearchStocks(ctx, query, limit)
	if err != nil {
		return nil, errors.InternalError("Failed to search stocks").WithDetails(err.Error())
	}

	return results, nil
}

func (s *Service) GetStockMetadata(ctx context.Context, symbol string) (*dto.StockMetadataResponse, error) {
	symbol = normalizeSymbol(symbol)
	key := "metadata:" + symbol

	// use singleflight to prevent duplicate requests for the same symbol (during high concurrency scenarios)
	v, err, _ := s.requestGroup.Do(key, func() (interface{}, error) {
		return s.getStockMetadataInternal(ctx, symbol)
	})
	if err != nil {
		return nil, err
	}

	stockMetadata, ok := v.(*dto.StockMetadataResponse)
	if !ok {
		return nil, errors.InternalError("Type assertion failed").
			WithDetails("Failed to assert type to StockMetadataResponse")
	}

	return stockMetadata, nil
}

func (s *Service) getStockMetadataInternal(ctx context.Context, symbol string) (*dto.StockMetadataResponse, error) {
	if !isValidSymbol(symbol) {
		return nil, errors.BadRequestError("Invalid symbol").
			WithDetails("The provided symbol is empty")
	}

	// check postgresql database
	stockMetadata, err := s.db.GetQueries().GetStockMetadataBySymbol(ctx, symbol)
	if err == nil {
		return utils.ConvertStockMetadataToDTO(stockMetadata), nil // if it exists, return it
	}

	providerMetadata, err := s.getStockMetadataFromProvider(ctx, symbol)
	if err != nil {
		return nil, errors.InternalError("Could not fetch stock metadata").
			WithDetails(err.Error())
	}

	return providerMetadata, nil
}

func (s *Service) GetStockQuote(ctx context.Context, symbol string) (*dto.StockQuoteResponse, error) {
	symbol = normalizeSymbol(symbol)
	key := "quote:" + symbol

	// use singleflight to prevent duplicate requests for the same symbol (during high concurrency scenarios)
	v, err, _ := s.requestGroup.Do(key, func() (interface{}, error) {
		return s.getStockQuoteInternal(ctx, symbol)
	})
	if err != nil {
		return nil, err
	}

	stockQuote, ok := v.(*dto.StockQuoteResponse)
	if !ok {
		return nil, errors.InternalError("Type assertion failed").
			WithDetails("Failed to assert type to StockQuoteResponse")
	}

	return stockQuote, nil
}

func (s *Service) getStockQuoteInternal(ctx context.Context, symbol string) (*dto.StockQuoteResponse, error) {
	if !isValidSymbol(symbol) {
		return nil, errors.BadRequestError("Invalid symbol").
			WithDetails("The provided symbol is empty")
	}

	// first, check redis cache
	cached, err := s.redis.Get(ctx, symbol)
	if err == nil && cached != "" {
		var stock dto.StockQuoteResponse
		if err := json.Unmarshal([]byte(cached), &stock); err == nil && stock.Currency != "" && s.quoteIsFresh(stock.FetchedAt) {
			return &stock, nil
		}
	}

	// PostgreSQL is a recovery cache, not the source of truth for current prices.
	// Only return a database quote while it is still inside the freshness window.
	stockQuote, err := s.db.GetQueries().GetStockQuoteBySymbol(ctx, symbol)
	if err == nil && s.quoteIsFresh(stockQuote.UpdatedAt.Time) {
		stock := utils.ConvertStockQuoteToDTO(stockQuote)
		metadata, metadataErr := s.db.GetQueries().GetStockMetadataBySymbol(ctx, symbol)
		if metadataErr == nil && metadata.Currency != "" {
			stock.Currency = metadata.Currency

			s.cacheQuote(ctx, symbol, &stock)

			return &stock, nil
		}
	}

	providerQuote, err := s.marketData.GetStockQuote(ctx, symbol)
	if err != nil {
		return nil, errors.InternalError("Failed to fetch stock quote").
			WithDetails(err.Error())
	}
	providerQuote.Symbol = normalizeSymbol(providerQuote.Symbol)
	if providerQuote.Symbol != symbol {
		return nil, errors.InternalError("Failed to fetch stock quote").
			WithDetails("The market data provider returned a different symbol")
	}
	if providerQuote.Currency == "" {
		metadata, metadataErr := s.GetStockMetadata(ctx, symbol)
		if metadataErr != nil || metadata.Currency == "" {
			return nil, errors.InternalError("Failed to determine quote currency").
				WithDetails("The market data provider returned a quote without a currency")
		}
		providerQuote.Currency = metadata.Currency
	}

	params := generated.InsertOrUpdateStockQuoteParams{
		Symbol:             providerQuote.Symbol,
		LastPrice:          providerQuote.LastPrice,
		OpenPrice:          providerQuote.OpenPrice,
		PreviousClosePrice: providerQuote.PreviousClose,
		DayHigh:            providerQuote.DayHigh,
		DayLow:             providerQuote.DayLow,
		YearHigh:           providerQuote.YearHigh,
		YearLow:            providerQuote.YearLow,
		Volume:             providerQuote.Volume,
		MarketCap:          providerQuote.MarketCap,
		PriceChange:        providerQuote.Change,
		PriceChangePct:     providerQuote.ChangePct,
		Source:             providerQuote.Source,
		AsOf:               pgtype.Timestamptz{Time: providerQuote.AsOf, Valid: true},
		MarketState:        providerQuote.MarketState,
	}

	// before storing stock quote in database, ensure stock metadata exists
	_, err = s.db.GetQueries().GetStockMetadataBySymbol(ctx, symbol)
	if err != nil {
		_, err = s.getStockMetadataFromProvider(ctx, symbol)
		if err != nil {
			log.Println("Warning: Failed to populate stock metadata: " + err.Error())
		}
	}

	// populate postgresql table with stock quote (or update if it already exists)
	_, err = s.db.GetQueries().InsertOrUpdateStockQuote(ctx, params)
	if err != nil {
		log.Println("Warning: Failed to store stock quote in database: " + err.Error())
	}

	s.cacheQuote(ctx, symbol, providerQuote)

	return providerQuote, nil
}

func (s *Service) GetStockQuoteBatch(ctx context.Context, symbols []string) ([]*dto.StockQuoteResponse, error) {
	if len(symbols) == 0 {
		return nil, errors.BadRequestError("Invalid symbols").
			WithDetails("The provided symbols list is empty")
	}

	normalizedSymbols := make([]string, len(symbols))
	for index, symbol := range symbols {
		normalizedSymbols[index] = normalizeSymbol(symbol)
		if !isValidSymbol(normalizedSymbols[index]) {
			return nil, errors.BadRequestError("Invalid symbol").
				WithDetails("The provided symbol " + normalizedSymbols[index] + " is invalid")
		}
	}

	quotes := make([]*dto.StockQuoteResponse, len(normalizedSymbols))
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(8)
	var fetchErrors []error
	var fetchErrorsMu sync.Mutex

	for index, symbol := range normalizedSymbols {
		group.Go(func() error {
			quote, err := s.GetStockQuote(groupCtx, symbol)
			if err != nil {
				fetchErrorsMu.Lock()
				fetchErrors = append(fetchErrors, fmt.Errorf("fetch quote for %s: %w", symbol, err))
				fetchErrorsMu.Unlock()
				return nil
			}

			quotes[index] = quote
			return nil
		})
	}

	_ = group.Wait()

	availableQuotes := make([]*dto.StockQuoteResponse, 0, len(quotes))
	for _, quote := range quotes {
		if quote != nil {
			availableQuotes = append(availableQuotes, quote)
		}
	}
	if len(availableQuotes) == 0 && len(fetchErrors) > 0 {
		return nil, fetchErrors[0]
	}
	for _, err := range fetchErrors {
		log.Printf("Warning: %v", err)
	}

	return availableQuotes, nil
}

func (s *Service) GetStockHistoricalData(ctx context.Context, symbol string, period string) ([]dto.StockHistoricalDataResponse, error) {
	symbol = normalizeSymbol(symbol)
	// check if symbol exists
	if !isValidSymbol(symbol) {
		return nil, errors.BadRequestError("Invalid symbol").
			WithDetails("The provided symbol is empty")
	}

	// check if period is valid and get date range
	from, to := utils.GetDateRangeFromPeriod(period)
	if from == "" || to == "" {
		return nil, errors.BadRequestError("Invalid period").
			WithDetails("The provided period is not valid")
	}

	// "convert" inputs to pgtype (for database queries)
	sym := pgtype.Text{String: symbol, Valid: true}
	fromDate := pgtype.Date{Time: utils.ParseDate(from), Valid: true}
	toDate := pgtype.Date{Time: utils.ParseDate(to), Valid: true}

	params := generated.GetStockHistoricalDataBySymbolAndDateRangeParams{
		Symbol: sym,
		Date:   fromDate,
		Date_2: toDate,
	}

	// try get from database first
	historicalData, err := s.db.GetQueries().GetStockHistoricalDataBySymbolAndDateRange(ctx, params)
	if err == nil && len(historicalData) > 0 {
		// check if we have sufficient data for the requested period
		if utils.HasCompleteDateRange(historicalData, fromDate.Time, toDate.Time, period) {
			var result []dto.StockHistoricalDataResponse
			for _, data := range historicalData {
				result = append(result, dto.StockHistoricalDataResponse{
					Symbol:     data.Symbol.String,
					Date:       data.Date.Time.Format("2006-01-02"),
					OpenPrice:  data.OpenPrice,
					HighPrice:  data.HighPrice,
					LowPrice:   data.LowPrice,
					ClosePrice: data.ClosePrice,
					Volume:     data.Volume,
					Change:     data.PriceChange,
					ChangePct:  data.PriceChangePct,
				})
			}

			log.Printf("Using cached historical data from database for %s", period)
			return result, nil
		} else {
			log.Printf("Insufficient cached data for period %s, fetching from API", period)
		}
	}

	providerHistory, err := s.marketData.GetStockHistoricalData(ctx, symbol, from, to)
	if err != nil {
		return nil, errors.InternalError("Failed to fetch historical stock data").
			WithDetails(err.Error())
	}

	// check if stock metadata exists before storing historical data
	_, err = s.db.GetQueries().GetStockMetadataBySymbol(ctx, symbol)
	if err != nil {
		_, err = s.getStockMetadataFromProvider(ctx, symbol)
		if err != nil {
			log.Println("Warning: Failed to populate stock metadata: " + err.Error())
		}
	}

	// store historical data in database
	for _, data := range providerHistory {
		sym := pgtype.Text{String: data.Symbol, Valid: true}
		date := pgtype.Date{Time: utils.ParseDate(data.Date), Valid: true}

		params := generated.InsertStockHistoricalDataParams{
			Symbol:         sym,
			Date:           date,
			OpenPrice:      data.OpenPrice,
			HighPrice:      data.HighPrice,
			LowPrice:       data.LowPrice,
			ClosePrice:     data.ClosePrice,
			Volume:         data.Volume,
			PriceChange:    data.Change,
			PriceChangePct: data.ChangePct,
		}

		_, err := s.db.GetQueries().InsertStockHistoricalData(ctx, params)
		if err != nil {
			log.Println("Warning: Failed to store historical stock data: " + err.Error())
		}
	}

	return providerHistory, nil
}

func (s *Service) getStockMetadataFromProvider(ctx context.Context, symbol string) (*dto.StockMetadataResponse, error) {
	providerMetadata, err := s.marketData.GetStockMetadata(ctx, symbol)
	if err != nil {
		log.Printf("Error fetching stock metadata for symbol %s: %v\n", symbol, err)
		return nil, errors.InternalError("Failed to fetch stock metadata").
			WithDetails(err.Error())
	}
	providerMetadata.Symbol = normalizeSymbol(providerMetadata.Symbol)
	if providerMetadata.Symbol != symbol || providerMetadata.Currency == "" || providerMetadata.InstrumentType == "" {
		return nil, errors.InternalError("Failed to fetch stock metadata").
			WithDetails("The market data provider returned incomplete metadata")
	}

	params := generated.CreateStockMetadataParams{
		Symbol:           providerMetadata.Symbol,
		Name:             providerMetadata.Name,
		Currency:         providerMetadata.Currency,
		Exchange:         providerMetadata.Exchange,
		ExchangeFullName: providerMetadata.ExchangeFullName,
		InstrumentType:   providerMetadata.InstrumentType,
	}

	_, err = s.db.GetQueries().CreateStockMetadata(ctx, params)
	if err != nil {
		log.Println("Warning: Failed to store stock metadata: " + err.Error())
	}

	return providerMetadata, nil
}

func normalizeSymbol(symbol string) string {
	return strings.ToUpper(strings.TrimSpace(symbol))
}

func isValidSymbol(symbol string) bool {
	return len(symbol) > 0 && len(symbol) <= 32 && !strings.ContainsAny(symbol, " \t\r\n")
}

func (s *Service) quoteIsFresh(fetchedAt time.Time) bool {
	if fetchedAt.IsZero() {
		return false
	}

	age := time.Since(fetchedAt)
	return age >= -time.Minute && age <= s.quoteTTL
}

func (s *Service) cacheQuote(ctx context.Context, symbol string, quote *dto.StockQuoteResponse) {
	data, err := json.Marshal(quote)
	if err != nil {
		log.Printf("Warning: Failed to encode %s quote for cache: %v", symbol, err)
		return
	}
	if err := s.redis.Set(ctx, symbol, string(data)); err != nil {
		log.Printf("Warning: Failed to cache %s quote: %v", symbol, err)
	}
}
