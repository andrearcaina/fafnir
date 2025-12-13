package utils

import (
	"fafnir/stock-service/internal/db/generated"
	"fafnir/stock-service/internal/dto"
)

func ConvertStockQuoteToDTO(dbQuote generated.StockQuote) dto.StockQuoteResponse {
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

func ConvertStockMetadataToDTO(dbMetadata generated.StockMetadatum) *dto.StockMetadataResponse {
	return &dto.StockMetadataResponse{
		Symbol:           dbMetadata.Symbol,
		Name:             dbMetadata.Name,
		Currency:         dbMetadata.Currency,
		Exchange:         dbMetadata.Exchange,
		ExchangeFullName: dbMetadata.ExchangeFullName,
	}
}

func ConvertMapToSlice(m map[string]*dto.StockQuoteResponse, keys []string) []*dto.StockQuoteResponse {
	result := make([]*dto.StockQuoteResponse, 0, len(keys))
	for _, key := range keys {
		if val, exists := m[key]; exists {
			result = append(result, val)
		}
	}
	return result
}
