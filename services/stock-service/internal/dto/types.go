package dto

type StockQuoteResponse struct {
	Symbol        string  `json:"symbol"`
	LastPrice     float64 `json:"price"`
	OpenPrice     float64 `json:"open"`
	PreviousClose float64 `json:"previousClose"`
	DayLow        float64 `json:"dayLow"`
	DayHigh       float64 `json:"dayHigh"`
	YearLow       float64 `json:"yearLow"`
	YearHigh      float64 `json:"yearHigh"`
	Volume        int64   `json:"volume"`
	MarketCap     int64   `json:"marketCap"`
	Change        float64 `json:"change"`
	ChangePct     float64 `json:"changePercentage"`
}

type StockMetadataResponse struct {
	Symbol           string `json:"symbol"`
	Name             string `json:"name"`
	Currency         string `json:"currency"`
	Exchange         string `json:"exchange"`
	ExchangeFullName string `json:"exchangeFullName"`
}
