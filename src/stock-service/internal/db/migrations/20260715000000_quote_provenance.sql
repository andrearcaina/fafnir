-- +goose Up
-- +goose StatementBegin
ALTER TABLE stock_quote
    ADD COLUMN source VARCHAR(32) NOT NULL DEFAULT 'unknown',
    ADD COLUMN as_of TIMESTAMPTZ,
    ADD COLUMN market_state VARCHAR(32) NOT NULL DEFAULT '';

UPDATE stock_quote
SET as_of = updated_at;

ALTER TABLE stock_quote
    ALTER COLUMN as_of SET NOT NULL,
    ALTER COLUMN as_of SET DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE stock_quote
    DROP COLUMN market_state,
    DROP COLUMN as_of,
    DROP COLUMN source;
-- +goose StatementEnd
