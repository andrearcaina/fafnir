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
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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

func (s *Service) GetStockQuote(ctx context.Context, symbol string) (*dto.StockQuoteResponse, error) {
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

func (s *Service) GetStockHistoricalData(ctx context.Context, symbol string, period string) ([]dto.StockHistoricalDataResponse, error) {
	// check if symbol exists
	if symbol == "" {
		return nil, errors.BadRequestError("Invalid symbol").
			WithDetails("The provided symbol is empty")
	}

	// check if period is valid and get date range
	from, to := getDateRangeFromPeriod(period)
	if from == "" || to == "" {
		return nil, errors.BadRequestError("Invalid period").
			WithDetails("The provided period is not valid")
	}

	sym := pgtype.Text{String: symbol, Valid: true}
	fromDate := pgtype.Date{Time: parseDate(from), Valid: true}
	toDate := pgtype.Date{Time: parseDate(to), Valid: true}

	params := generated.GetStockHistoricalDataBySymbolAndDateRangeParams{
		Symbol: sym,
		Date:   fromDate,
		Date_2: toDate,
	}

	// try get from database first
	historicalData, err := s.db.GetQueries().GetStockHistoricalDataBySymbolAndDateRange(ctx, params)
	if err == nil && len(historicalData) > 0 {
		// check if we have sufficient data for the requested period
		if hasCompleteDateRange(historicalData, fromDate.Time, toDate.Time, period) {
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
		date := pgtype.Date{Time: parseDate(data.Date), Valid: true}

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

func getDateRangeFromPeriod(period string) (string, string) {
	validPeriods := map[string]bool{
		"1D":  true,
		"1W":  true,
		"1M":  true,
		"3M":  true,
		"6M":  true,
		"1Y":  true,
		"2Y":  true,
		"5Y":  true,
		"MAX": true,
	}

	if !validPeriods[period] {
		return "", ""
	}

	now := time.Now().UTC()
	var start time.Time

	switch period {
	case "1D":
		start = now.AddDate(0, 0, -1)
	case "1W":
		start = now.AddDate(0, 0, -7)
	case "1M":
		start = now.AddDate(0, -1, 0)
	case "3M":
		start = now.AddDate(0, -3, 0)
	case "6M":
		start = now.AddDate(0, -6, 0)
	case "1Y":
		start = now.AddDate(-1, 0, 0)
	case "2Y":
		start = now.AddDate(-2, 0, 0)
	case "5Y":
		start = now.AddDate(-5, 0, 0)
	case "MAX":
		start = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	return start.Format("2006-01-02"), now.Format("2006-01-02")
}

func parseDate(dateStr string) time.Time {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Printf("Warning: Failed to parse date '%s': %v", dateStr, err)
		return time.Now().UTC()
	}
	return date
}

func hasCompleteDateRange(historicalData []generated.StockHistoricalDatum, fromDate time.Time, toDate time.Time, period string) bool {
	if len(historicalData) == 0 {
		return false
	}

	// get the actual date range from the data
	firstDataDate := historicalData[0].Date.Time
	lastDataDate := historicalData[len(historicalData)-1].Date.Time

	// check if the data covers the requested date range
	// Allow for some tolerance (weekends, holidays)
	dateTolerance := 3 * 24 * time.Hour // 3 days tolerance

	// check if the data starts close enough to the requested start date
	if firstDataDate.After(fromDate.Add(dateTolerance)) {
		log.Printf("Data starts too late: have %s, need %s", firstDataDate.Format("2006-01-02"), fromDate.Format("2006-01-02"))
		return false
	}

	// check if the data ends close enough to the requested end date
	if lastDataDate.Before(toDate.Add(-dateTolerance)) {
		log.Printf("Data ends too early: have %s, need %s", lastDataDate.Format("2006-01-02"), toDate.Format("2006-01-02"))
		return false
	}

	// check if we have reasonable data coverage (at least 70% of expected average)
	expectedDataPoints := getExpectedDataPointsForPeriod(period)
	minRequiredDataPoints := int(float64(expectedDataPoints) * 0.7)

	if len(historicalData) < minRequiredDataPoints {
		log.Printf("Insufficient data coverage: have %d, need at least %d (70%% of %d expected) for period %s",
			len(historicalData), minRequiredDataPoints, expectedDataPoints, period)
		return false
	}

	log.Printf("Data coverage sufficient: have %d, expected ~%d for period %s",
		len(historicalData), expectedDataPoints, period)
	return true
}

func getExpectedDataPointsForPeriod(period string) int {
	// returns the typical/average number of trading days expected for each period
	// markets are typically open ~252 days per year (5 days/week, minus holidays)
	switch period {
	case "1D":
		return 1
	case "1W":
		return 5 // 5 trading days per week
	case "1M":
		return 21 // ~21 trading days per month (252/12)
	case "3M":
		return 63 // ~63 trading days per quarter (252/4)
	case "6M":
		return 126 // ~126 trading days per half year (252/2)
	case "1Y":
		return 252 // ~252 trading days per year
	case "2Y":
		return 504 // ~504 trading days per 2 years
	case "5Y":
		return 1260 // ~1260 trading days per 5 years
	case "MAX":
		return 5000 // Conservative estimate for max historical data
	default:
		return 1
	}
}
