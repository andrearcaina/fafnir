-- +goose Up
-- +goose StatementBegin
-- stores stock information
CREATE TABLE stock_metadata (
    symbol VARCHAR(10) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    type VARCHAR(50) NOT NULL,
    sector VARCHAR(100) NOT NULL
);

-- This table stores the latest stock quote information (data gets overwritten if updated_at > 60 seconds, acts like caching)
CREATE TABLE stock_quote (
     symbol VARCHAR(10) REFERENCES stock_metadata(symbol) ON DELETE CASCADE,
     last_price NUMERIC NOT NULL,
     price_change NUMERIC NOT NULL,
     price_change_pct NUMERIC NOT NULL,
     volume BIGINT NOT NULL,
     market_cap BIGINT NOT NULL,
     pe_ratio NUMERIC NOT NULL,
     dividend_yield NUMERIC NOT NULL,
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS stock_quote;
DROP TABLE IF EXISTS stock_metadata;
-- +goose StatementEnd
