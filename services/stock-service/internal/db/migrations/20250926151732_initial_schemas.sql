-- +goose Up
-- +goose StatementBegin
-- stores stock information
CREATE TABLE stock_metadata (
    symbol VARCHAR(10) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    exchange_full_name VARCHAR(100) NOT NULL,
    currency VARCHAR(10) NOT NULL
);

-- This table stores the latest stock quote information (data gets overwritten if updated_at > 60 seconds, acts like caching)
CREATE TABLE stock_quote (
     symbol VARCHAR(10) REFERENCES stock_metadata(symbol) ON DELETE CASCADE,
     open_price FLOAT NOT NULL,
     last_price FLOAT NOT NULL,
     previous_close_price FLOAT NOT NULL,
     price_change FLOAT NOT NULL,
     price_change_pct FLOAT NOT NULL,
     volume BIGINT NOT NULL,
     market_cap FLOAT NOT NULL,
     day_low FLOAT NOT NULL,
     day_high FLOAT NOT NULL,
     year_low FLOAT NOT NULL,
     year_high FLOAT NOT NULL,
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     PRIMARY KEY (symbol)
);
-- +goose StatementEnd

CREATE TABLE stock_historical_data (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(10) REFERENCES stock_metadata(symbol) ON DELETE CASCADE,
    date DATE NOT NULL,
    open_price FLOAT NOT NULL,
    close_price FLOAT NOT NULL,
    high_price FLOAT NOT NULL,
    low_price FLOAT NOT NULL,
    volume BIGINT NOT NULL,
    price_change FLOAT NOT NULL,
    price_change_pct FLOAT NOT NULL,
    UNIQUE(symbol, date)
);

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS stock_quote;
DROP TABLE IF EXISTS stock_historical_data;
DROP TABLE IF EXISTS stock_metadata;
-- +goose StatementEnd
