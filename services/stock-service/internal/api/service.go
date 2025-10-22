package api

import (
	"context"
	"encoding/json"
	"fafnir/shared/pkg/errors"
	"fafnir/stock-service/internal/db"
	"fafnir/stock-service/internal/db/generated"
	"fafnir/stock-service/internal/dto"
	"fafnir/stock-service/internal/fmp"
	"fafnir/stock-service/internal/redis"
	"log"
)

type Service struct {
	db    *db.Database
	redis *redis.Cache
	fmp   *fmp.Client
}

func NewStockService(database *db.Database, redis *redis.Cache, fmp *fmp.Client) *Service {
	return &Service{
		db:    database,
		redis: redis,
		fmp:   fmp,
	}
}

func (s *Service) SearchStockQuote(ctx context.Context, symbol string) (*dto.StockQuoteResponse, error) {
	if symbol == "" {
		return nil, errors.BadRequestError("Invalid symbol").
			WithDetails("The provided symbol is empty")
	}

	// first, check redis cache
	cached, err := s.redis.Get(ctx, symbol)
	if err == nil && cached != "" {
		var stock dto.StockQuoteResponse
		if err := json.Unmarshal([]byte(cached), &stock); err == nil {
			return &stock, nil
		}
	}

	// if not found in cache, check postgresql database
	stockQuote, err := s.db.GetQueries().GetStockQuoteBySymbol(ctx, symbol)
	// if found in database, populate redis cache and return
	if err == nil {
		stock := convertStockQuoteToDTO(stockQuote)

		// populate redis cache for future requests
		data, _ := json.Marshal(stock)
		err := s.redis.Set(ctx, symbol, string(data))
		if err != nil {
			log.Println("Warning: Failed to cache stock quote from database: " + err.Error()) // don't fail the request if caching fails
		}

		return &stock, nil
	}

	// if not found in redis + database, call fmp client to get stock quote
	fmpStockQuote, err := s.fmp.GetStockQuote(symbol)
	if err != nil {
		return nil, errors.InternalError("Failed to fetch stock quote").
			WithDetails("Error calling FMP API: " + err.Error())
	}

	params := generated.InsertOrUpdateStockQuoteParams{
		Symbol:             fmpStockQuote.Symbol,
		LastPrice:          fmpStockQuote.LastPrice,
		OpenPrice:          fmpStockQuote.OpenPrice,
		PreviousClosePrice: fmpStockQuote.PreviousClose,
		DayHigh:            fmpStockQuote.DayHigh,
		DayLow:             fmpStockQuote.DayLow,
		YearHigh:           fmpStockQuote.YearHigh,
		YearLow:            fmpStockQuote.YearLow,
		Volume:             fmpStockQuote.Volume,
		MarketCap:          fmpStockQuote.MarketCap,
		PriceChange:        fmpStockQuote.Change,
		PriceChangePct:     fmpStockQuote.ChangePct,
	}

	// before storing stock quote in database, ensure stock metadata exists
	_, err = s.db.GetQueries().GetStockMetadataBySymbol(ctx, symbol)
	if err != nil {
		_, err = s.getStockMetadataFromFMP(ctx, symbol) // store stock metadata from FMP
		if err != nil {
			log.Println("Warning: Failed to populate stock metadata from FMP: " + err.Error())
		}
	}

	// populate postgresql table with stock quote (or update if it already exists)
	_, err = s.db.GetQueries().InsertOrUpdateStockQuote(ctx, params)
	if err != nil {
		log.Println("Warning: Failed to store stock quote from FMP to database: " + err.Error())
	}

	data, _ := json.Marshal(fmpStockQuote)
	err = s.redis.Set(ctx, symbol, string(data))
	if err != nil {
		log.Println("Warning: Failed to cache stock quote from FMP: " + err.Error())
	}

	return fmpStockQuote, nil
}

func (s *Service) GetStockMetadata(ctx context.Context, symbol string) (*dto.StockMetadataResponse, error) {
	if symbol == "" {
		return nil, errors.BadRequestError("Invalid symbol").
			WithDetails("The provided symbol is empty")
	}

	// check postgresql database
	stockMetadata, err := s.db.GetQueries().GetStockMetadataBySymbol(ctx, symbol)
	if err == nil {
		return convertStockMetadataToDTO(stockMetadata), nil // if it exists, return it
	}

	fmpStockMetadata, err := s.getStockMetadataFromFMP(ctx, symbol)
	if err != nil {
		return nil, err
	}

	return fmpStockMetadata, nil
}

func (s *Service) getStockMetadataFromFMP(ctx context.Context, symbol string) (*dto.StockMetadataResponse, error) {
	fmpStockMetadata, err := s.fmp.GetStockMetadata(symbol)
	if err != nil {
		log.Printf("Error fetching stock metadata from FMP for symbol %s: %v\n", symbol, err)
		return nil, errors.InternalError("Failed to fetch stock metadata").
			WithDetails("Error calling FMP API: " + err.Error())
	}

	params := generated.CreateStockMetadataParams{
		Symbol:           fmpStockMetadata.Symbol,
		Name:             fmpStockMetadata.Name,
		Currency:         fmpStockMetadata.Currency,
		Exchange:         fmpStockMetadata.Exchange,
		ExchangeFullName: fmpStockMetadata.ExchangeFullName,
	}

	data, err := s.db.GetQueries().CreateStockMetadata(ctx, params)
	if err != nil {
		log.Println("Warning: Failed to store stock metadata from FMP to database: " + err.Error())
	}

	log.Printf("Successfully stored stock metadata: %+v\n", data)

	return fmpStockMetadata, nil
}

func convertStockQuoteToDTO(dbQuote generated.StockQuote) dto.StockQuoteResponse {
	return dto.StockQuoteResponse{
		Symbol:        dbQuote.Symbol,
		LastPrice:     dbQuote.LastPrice,
		OpenPrice:     dbQuote.OpenPrice,
		PreviousClose: dbQuote.PreviousClosePrice,
		DayHigh:       dbQuote.DayHigh,
		DayLow:        dbQuote.DayLow,
		YearHigh:      dbQuote.YearHigh,
		YearLow:       dbQuote.YearLow,
		Volume:        dbQuote.Volume,
		MarketCap:     dbQuote.MarketCap,
		Change:        dbQuote.PriceChange,
		ChangePct:     dbQuote.PriceChangePct,
	}
}

func convertStockMetadataToDTO(dbMetadata generated.StockMetadatum) *dto.StockMetadataResponse {
	return &dto.StockMetadataResponse{
		Symbol:           dbMetadata.Symbol,
		Name:             dbMetadata.Name,
		Currency:         dbMetadata.Currency,
		Exchange:         dbMetadata.Exchange,
		ExchangeFullName: dbMetadata.ExchangeFullName,
	}
}
