package stocks

var supportedSymbols = []string{
	"AAPL", "AAL", "ABBV", "ADBE", "AMD", "AMZN", "ATVI", "BA", "BABA", "BAC",
	"BIDU", "BILI", "C", "CARR", "CCL", "COIN", "COST", "CPRX", "CSCO", "CVX",
	"DAL", "DIS", "DOCU", "ET", "ETSY", "F", "FDX", "FUBO", "GE", "GM",
	"GOOGL", "GS", "HCA", "HOOD", "INTC", "JNJ", "JPM", "KO", "LCID", "LMT",
	"META", "MGM", "MRNA", "MRO", "MSFT", "NFLX", "NIO", "NKE", "NOK", "NVDA",
	"PEP", "PFE", "PINS", "PLTR", "PYPL", "RBLX", "RIOT", "RIVN", "RKT", "ROKU",
	"SBUX", "SHOP", "SIRI", "SNAP", "SOFI", "SONY", "SPY", "SPYG", "SQ", "T",
	"TGT", "TLRY", "TSLA", "TSM", "TWTR", "UAL", "UBER", "UNH", "V", "VIAC",
	"VWO", "VZ", "WBA", "WFC", "WMT", "XOM", "ZM",
}

var supportedSymbolSet = func() map[string]struct{} {
	symbols := make(map[string]struct{}, len(supportedSymbols))
	for _, symbol := range supportedSymbols {
		symbols[symbol] = struct{}{}
	}
	return symbols
}()

func IsSupported(symbol string) bool {
	_, ok := supportedSymbolSet[symbol]
	return ok
}

func SupportedSymbols() []string {
	return append([]string(nil), supportedSymbols...)
}
