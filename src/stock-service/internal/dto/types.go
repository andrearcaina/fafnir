package dto

import "time"

type StockQuoteResponse struct {
	Symbol        string    `json:"symbol"`
	LastPrice     float64   `json:"price"`
	OpenPrice     float64   `json:"open"`
	PreviousClose float64   `json:"previousClose"`
	DayLow        float64   `json:"dayLow"`
	DayHigh       float64   `json:"dayHigh"`
	YearLow       float64   `json:"yearLow"`
	YearHigh      float64   `json:"yearHigh"`
	Volume        int64     `json:"volume"`
	MarketCap     float64   `json:"marketCap"`
	Change        float64   `json:"change"`
	ChangePct     float64   `json:"changePercentage"`
	Source        string    `json:"source"`
	AsOf          time.Time `json:"asOf"`
	FetchedAt     time.Time `json:"fetchedAt"`
	MarketState   string    `json:"marketState,omitempty"`
	Currency      string    `json:"currency"`
}

type StockMetadataResponse struct {
	Symbol           string `json:"symbol"`
	Name             string `json:"name"`
	Currency         string `json:"currency"`
	Exchange         string `json:"exchange"`
	ExchangeFullName string `json:"exchangeFullName"`
	InstrumentType   string `json:"instrumentType"`
}

type StockSearchResult struct {
	Symbol           string `json:"symbol"`
	Name             string `json:"name"`
	Exchange         string `json:"exchange"`
	ExchangeFullName string `json:"exchangeFullName"`
	InstrumentType   string `json:"instrumentType"`
}

type StockHistoricalDataResponse struct {
	Symbol     string  `json:"symbol"`
	Date       string  `json:"date"`
	OpenPrice  float64 `json:"open"`
	HighPrice  float64 `json:"high"`
	LowPrice   float64 `json:"low"`
	ClosePrice float64 `json:"close"`
	Volume     int64   `json:"volume"`
	Change     float64 `json:"change"`
	ChangePct  float64 `json:"changePercent"`
}
