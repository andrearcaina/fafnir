SYMBOLS = ["AAPL", "MSFT", "TSLA", "AMZN", "GOOGL", "META", "NVDA", "NFLX"]
PERIODS = ["1D", "1W", "1M", "3M", "6M"]  # no need for more than 1Y in load tests

GET_STOCK_QUOTE = """
query GetStockQuote($symbol: String!) {
  getStockQuote(symbol: $symbol) {
    data {
      symbol
      price
      open
      previousClose
      priceChange
      priceChangePercent
      volume
      marketCap
      dayLow
      dayHigh
      yearHigh
      yearLow
    }
    code
  }
}
"""

GET_STOCK_QUOTE_BATCH = """
query GetStockQuoteBatch($symbols: [String!]!) {
  getStockQuoteBatch(symbols: $symbols) {
    code
    data {
      symbol
      price
      open
      previousClose
      priceChange
      priceChangePercent
      volume
      marketCap
      dayLow
      dayHigh
      yearHigh
      yearLow
    }
  }
}
"""

GET_STOCK_METADATA = """
query GetStockMetadata($symbol: String!) {
  getStockMetadata(symbol: $symbol) {
    data {
      symbol
      name
      exchange
      exchangeFullName
      currency
    }
    code
  }
}
"""

GET_STOCK_HISTORY = """
query GetStockHistory($symbol: String!, $period: String!) {
  getStockHistoricalData(symbol: $symbol, period: $period) {
    data {
      date
      open
      high
      low
      close
      volume
    }
    code
  }
}
"""
