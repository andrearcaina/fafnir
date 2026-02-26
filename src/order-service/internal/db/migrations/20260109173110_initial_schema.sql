-- +goose Up
-- +goose StatementBegin
CREATE TYPE order_side AS ENUM ('buy', 'sell');
CREATE TYPE order_type AS ENUM ('market', 'limit', 'stop', 'stop_limit');
CREATE TYPE order_status AS ENUM ('pending', 'partially_filled', 'filled', 'canceled', 'rejected');

-- this table tracks orders placed by users
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    side order_side NOT NULL,
    type order_type NOT NULL,
    status order_status NOT NULL DEFAULT 'pending',
    quantity NUMERIC(20, 6) NOT NULL CHECK (quantity > 0),
    filled_quantity NUMERIC(20, 6) NOT NULL DEFAULT 0,
    price NUMERIC(20, 6) CHECK (price > 0),
    stop_price NUMERIC(20, 6) CHECK (stop_price > 0),
    avg_fill_price NUMERIC(20, 6) DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- this table tracks individual fills for orders
CREATE TABLE orders_fill (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    fill_quantity NUMERIC(20, 6) NOT NULL CHECK (fill_quantity > 0),
    fill_price NUMERIC(20, 6) NOT NULL CHECK (fill_price > 0),
    filled_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- this "index" speeds up lookups of orders by user and status
CREATE INDEX idx_orders_user_status ON orders(user_id, status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders_fill;
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS order_type;
DROP TYPE IF EXISTS order_side;
-- +goose StatementEnd
