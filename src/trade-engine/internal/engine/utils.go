package engine

import portfoliopb "fafnir/shared/pb/portfolio"

func getCurrencyString(c portfoliopb.CurrencyType) string {
	switch c {
	case portfoliopb.CurrencyType_CURRENCY_TYPE_USD:
		return "USD"
	case portfoliopb.CurrencyType_CURRENCY_TYPE_CAD:
		return "CAD"
	default:
		return "USD" // default fallback
	}
}

func getExchangeRate(from string, to string) float64 {
	if from == to {
		return 1.0
	}
	// mock FX rates (VERY simplified)
	if from == "USD" && to == "CAD" {
		return 1.35
	}
	if from == "CAD" && to == "USD" {
		return 0.74
	}
	return 1.0 // default
}
