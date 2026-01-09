-- name: InsertOrderFilled :one
INSERT INTO orders_fill (order_id, fill_quantity, fill_price, filled_at)
VALUES ($1, $2, $3, $4)
RETURNING *;
