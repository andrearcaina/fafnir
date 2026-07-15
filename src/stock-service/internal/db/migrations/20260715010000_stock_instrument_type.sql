-- +goose Up
-- +goose StatementBegin
ALTER TABLE stock_metadata
    ADD COLUMN instrument_type VARCHAR(32) NOT NULL DEFAULT 'EQUITY';

UPDATE stock_metadata
SET instrument_type = 'ETF'
WHERE symbol IN ('SPY', 'SPYG', 'VWO');

ALTER TABLE stock_quote
    DROP CONSTRAINT stock_quote_symbol_fkey;

ALTER TABLE stock_historical_data
    DROP CONSTRAINT stock_historical_data_symbol_fkey;

ALTER TABLE stock_metadata
    ALTER COLUMN symbol TYPE VARCHAR(32);

ALTER TABLE stock_historical_data
    ALTER COLUMN symbol TYPE VARCHAR(32);

ALTER TABLE stock_quote
    ALTER COLUMN symbol TYPE VARCHAR(32);

ALTER TABLE stock_quote
    ADD CONSTRAINT stock_quote_symbol_fkey
    FOREIGN KEY (symbol) REFERENCES stock_metadata(symbol) ON DELETE CASCADE;

ALTER TABLE stock_historical_data
    ADD CONSTRAINT stock_historical_data_symbol_fkey
    FOREIGN KEY (symbol) REFERENCES stock_metadata(symbol) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE stock_metadata
    DROP COLUMN instrument_type;

ALTER TABLE stock_quote
    DROP CONSTRAINT stock_quote_symbol_fkey;

ALTER TABLE stock_historical_data
    DROP CONSTRAINT stock_historical_data_symbol_fkey;

ALTER TABLE stock_metadata
    ALTER COLUMN symbol TYPE VARCHAR(10);

ALTER TABLE stock_historical_data
    ALTER COLUMN symbol TYPE VARCHAR(10);

ALTER TABLE stock_quote
    ALTER COLUMN symbol TYPE VARCHAR(10);

ALTER TABLE stock_quote
    ADD CONSTRAINT stock_quote_symbol_fkey
    FOREIGN KEY (symbol) REFERENCES stock_metadata(symbol) ON DELETE CASCADE;

ALTER TABLE stock_historical_data
    ADD CONSTRAINT stock_historical_data_symbol_fkey
    FOREIGN KEY (symbol) REFERENCES stock_metadata(symbol) ON DELETE CASCADE;
-- +goose StatementEnd
