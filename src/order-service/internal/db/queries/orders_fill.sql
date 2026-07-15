-- name: InsertOrderFilled :exec
INSERT INTO orders_fill (order_id, fill_quantity, fill_price, filled_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (order_id) DO NOTHING;
