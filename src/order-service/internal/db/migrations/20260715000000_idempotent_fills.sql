-- +goose Up
-- +goose StatementBegin
DELETE FROM orders_fill duplicate
USING orders_fill original
WHERE duplicate.order_id = original.order_id
  AND (duplicate.filled_at, duplicate.id) > (original.filled_at, original.id);

CREATE UNIQUE INDEX orders_fill_order_id_unique ON orders_fill(order_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS orders_fill_order_id_unique;
-- +goose StatementEnd
