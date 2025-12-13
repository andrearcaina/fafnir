package utils

import (
	"fafnir/stock-service/internal/db/generated"
	"log"
	"time"
)

func GetDateRangeFromPeriod(period string) (string, string) {
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
		start = now.AddDate(0, 0, -1) // 1 day ago
	case "1W":
		start = now.AddDate(0, 0, -7) // 7 days ago
	case "1M":
		start = now.AddDate(0, -1, 0) // 1 month ago
	case "3M":
		start = now.AddDate(0, -3, 0) // 3 months ago
	case "6M":
		start = now.AddDate(0, -6, 0) // 6 months ago
	case "1Y":
		start = now.AddDate(-1, 0, 0) // 1 year ago
	case "2Y":
		start = now.AddDate(-2, 0, 0) // 2 years ago
	case "5Y":
		start = now.AddDate(-5, 0, 0) // 5 years ago
	case "MAX":
		start = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC) // earliest possible date
	}

	return start.Format("2006-01-02"), now.Format("2006-01-02")
}

func ParseDate(dateStr string) time.Time {
	date, err := time.Parse("2006-01-02", dateStr) // ISO 8601 format
	if err != nil {
		log.Printf("Warning: Failed to parse date '%s': %v", dateStr, err)
		return time.Now().UTC()
	}
	return date
}

func HasCompleteDateRange(historicalData []generated.StockHistoricalDatum, fromDate time.Time, toDate time.Time, period string) bool {
	if len(historicalData) == 0 {
		log.Printf("No historical data available")
		return false
	}

	// get the actual date range from the data
	firstDataDate := historicalData[0].Date.Time
	lastDataDate := historicalData[len(historicalData)-1].Date.Time

	// check if the data covers the requested date range
	// allow for some tolerance (weekends, holidays)
	dateTolerance := 3 * 24 * time.Hour // 3 days tolerance (basically on 1 weekend + 1 holiday)

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
	expectedDataPoints := getIntervalForPeriod(period)
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

func getIntervalForPeriod(period string) int {
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
