package api

import (
	"context"
	"encoding/json"
	"fafnir/shared/pkg/errors"
	"fafnir/shared/pkg/redis"
	"fafnir/stock-service/internal/db"
	"fafnir/stock-service/internal/db/generated"
	"fafnir/stock-service/internal/dto"
	"fafnir/stock-service/internal/fmp"
	"fafnir/stock-service/internal/utils"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/singleflight"
)

type Service struct {
	db           *db.Database
	redis        *redis.Cache
	fmp          *fmp.Client
	requestGroup singleflight.Group
}

func NewStockService(database *db.Database, redis *redis.Cache, fmp *fmp.Client) *Service {
	return &Service{
		db:    database,
		redis: redis,
		fmp:   fmp,
	}
}

func (s *Service) GetStockMetadata(ctx context.Context, symbol string) (*dto.StockMetadataResponse, error) {
	key := "metadata: " + symbol

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
	if symbol == "" || !utils.GetValidStocks()[symbol] {
		return nil, errors.BadRequestError("Invalid symbol").
			WithDetails("The provided symbol is empty")
	}

	// check postgresql database
	stockMetadata, err := s.db.GetQueries().GetStockMetadataBySymbol(ctx, symbol)
	if err == nil {
		return utils.ConvertStockMetadataToDTO(stockMetadata), nil // if it exists, return it
	}

	fmpStockMetadata, err := s.getStockMetadataFromFMP(ctx, symbol)
	if err != nil {
		return nil, errors.InternalError("Could not fetch stock metadata from FMP API").
			WithDetails("Error fetching stock metadata from FMP")
	}

	return fmpStockMetadata, nil
}

func (s *Service) GetStockQuote(ctx context.Context, symbol string) (*dto.StockQuoteResponse, error) {
	key := "quote: " + symbol

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
	if symbol == "" || !utils.GetValidStocks()[symbol] {
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
		stock := utils.ConvertStockQuoteToDTO(stockQuote)

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

func (s *Service) GetStockQuoteBatch(ctx context.Context, symbols []string) ([]*dto.StockQuoteResponse, error) {
	if len(symbols) == 0 {
		return nil, errors.BadRequestError("Invalid symbols").
			WithDetails("The provided symbols list is empty")
	}

	// check for invalid symbols
	for _, symbol := range symbols {
		if symbol == "" || !utils.GetValidStocks()[symbol] {
			return nil, errors.BadRequestError("Invalid symbol").
				WithDetails("The provided symbol " + symbol + " is invalid")
		}
	}

	result := make(map[string]*dto.StockQuoteResponse)
	missingSymbols := make([]string, 0)

	// first, try get from redis cache
	// use redis MGET for batch retrieval
	cachedResults, err := s.redis.MGet(ctx, symbols)
	if err == nil {
		for i, cached := range cachedResults {
			// for each symbol, check if it was found in cache
			symbol := symbols[i]

			// if found in cache, unmarshal and add to result
			if cachedStr, ok := cached.(string); ok {
				var stock dto.StockQuoteResponse
				if err := json.Unmarshal([]byte(cachedStr), &stock); err == nil {
					result[symbol] = &stock
					continue
				}
			}

			missingSymbols = append(missingSymbols, symbol)
		}
	} else {
		// if Redis fails entirely, fetch everything from DB
		missingSymbols = symbols
	}

	if len(missingSymbols) == 0 {
		return utils.ConvertMapToSlice(result, symbols), nil
	}

	// for missing symbols, try get from database
	for _, symbol := range missingSymbols {
		stockQuote, err := s.db.GetQueries().GetStockQuoteBySymbol(ctx, symbol)
		if err == nil {
			stock := utils.ConvertStockQuoteToDTO(stockQuote)
			result[symbol] = &stock

			// populate redis cache for future requests
			data, _ := json.Marshal(stock)
			err := s.redis.Set(ctx, symbol, string(data))
			if err != nil {
				log.Println("Warning: Failed to cache stock quote from database: " + err.Error())
			}
			continue
		}
	}

	// check which symbols are still missing
	stillMissingSymbols := make([]string, 0)
	for _, symbol := range missingSymbols {
		if _, exists := result[symbol]; !exists {
			stillMissingSymbols = append(stillMissingSymbols, symbol)
		}
	}

	// for still missing symbols, call fmp client
	for _, symbol := range stillMissingSymbols {
		fmpStockQuote, err := s.fmp.GetStockQuote(symbol)
		if err != nil {
			log.Println("Warning: Failed to fetch stock quote from FMP for symbol " + symbol + ": " + err.Error())
			continue
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

		result[symbol] = fmpStockQuote
	}

	return utils.ConvertMapToSlice(result, symbols), nil
}

func (s *Service) GetStockHistoricalData(ctx context.Context, symbol string, period string) ([]dto.StockHistoricalDataResponse, error) {
	// check if symbol exists
	if symbol == "" || !utils.GetValidStocks()[symbol] {
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

	// if not found in database, call fmp client
	fmpStockHistoricalData, err := s.fmp.GetStockHistoricalData(symbol, from, to)
	if err != nil {
		return nil, errors.InternalError("Failed to fetch historical stock data").
			WithDetails("Error calling FMP API: " + err.Error())
	}

	// check if stock metadata exists before storing historical data
	_, err = s.db.GetQueries().GetStockMetadataBySymbol(ctx, symbol)
	if err != nil {
		_, err = s.getStockMetadataFromFMP(ctx, symbol) // store stock metadata from FMP
		if err != nil {
			log.Println("Warning: Failed to populate stock metadata from FMP: " + err.Error())
		}
	}

	// store historical data in database
	for _, data := range fmpStockHistoricalData {
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
			log.Println("Warning: Failed to store historical stock data from FMP to database: " + err.Error())
		}
	}

	return fmpStockHistoricalData, nil
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

	_, err = s.db.GetQueries().CreateStockMetadata(ctx, params)
	if err != nil {
		log.Println("Warning: Failed to store stock metadata from FMP to database: " + err.Error())
	}

	return fmpStockMetadata, nil
}
