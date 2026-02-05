-- +goose Up
-- +goose StatementBegin
-- -- stores account information for users (like their balance)
CREATE TYPE account_type AS ENUM ('savings', 'investment', 'chequing');
CREATE TYPE currency_type AS ENUM ('USD', 'CAD');
CREATE TYPE transaction_type AS ENUM ('deposit', 'transfer_in', 'transfer_out', 'buy', 'sell');

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    account_number VARCHAR(20) NOT NULL UNIQUE, -- just for display
    account_type account_type NOT NULL,
    currency currency_type NOT NULL,
    balance NUMERIC(20, 6) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- a list of stocks a user holds
CREATE TABLE holdings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    symbol VARCHAR(10) NOT NULL, -- this must reference what is in stocks_metadata of stocks-service
    quantity NUMERIC(20,6) NOT NULL CHECK (quantity >= 0),
    avg_cost NUMERIC(20,6) NOT NULL CHECK (avg_cost >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- one entry per stock per account (this means we update quantity and avg_cost on buy/sell)
    UNIQUE (account_id, symbol)
);

-- stocks a user is watching
CREATE TABLE watchlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, symbol) -- same as above, one entry per stock per user
);

-- logs all changes to accounts and holdings (like a ledger or audit history)
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    transaction_type transaction_type NOT NULL, -- deposit, transfer, buy, sell
    amount NUMERIC(20,6) NOT NULL CHECK (amount >= 0),
    description TEXT NOT NULL, -- e.g., "Bought 10 shares of AAPL"
    reference_id UUID, -- references orders table ID from orders-service or other services
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit;
DROP TABLE IF EXISTS watchlists;
DROP TABLE IF EXISTS holdings;
DROP TABLE IF EXISTS accounts;
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS currency_type;
DROP TYPE IF EXISTS account_type;
-- +goose StatementEnd
