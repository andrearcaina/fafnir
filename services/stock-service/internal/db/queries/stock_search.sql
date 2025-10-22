-- name: GetStockMetadataBySymbol :one
SELECT * FROM stock_metadata WHERE symbol = $1;

-- name: SearchStockMetadataByName :many
SELECT * FROM stock_metadata WHERE name ILIKE '%' || $1 || '%' LIMIT $2 OFFSET $3;

-- name: GetStockQuoteBySymbol :one
SELECT * FROM stock_quote WHERE symbol = $1;

-- name: CreateStockMetadata :one
INSERT INTO stock_metadata (symbol, name, exchange, exchange_full_name, currency)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: InsertOrUpdateStockQuote :one
INSERT INTO stock_quote (
    symbol, open_price, last_price, previous_close_price, price_change, price_change_pct,
    volume, market_cap, day_low, day_high, year_low, year_high, updated_at
)
VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11, $12, NOW()
)
ON CONFLICT (symbol) DO UPDATE SET
    open_price = EXCLUDED.open_price,
    last_price = EXCLUDED.last_price,
    previous_close_price = EXCLUDED.previous_close_price,
    price_change = EXCLUDED.price_change,
    price_change_pct = EXCLUDED.price_change_pct,
    volume = EXCLUDED.volume,
    market_cap = EXCLUDED.market_cap,
    day_low = EXCLUDED.day_low,
    day_high = EXCLUDED.day_high,
    year_low = EXCLUDED.year_low,
    year_high = EXCLUDED.year_high,
    updated_at = NOW()
RETURNING *;