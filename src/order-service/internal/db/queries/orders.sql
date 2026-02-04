-- name: GetOrderById :one
SELECT * FROM orders
WHERE id = $1 LIMIT 1;

-- name: GetOrdersByUserId :many
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: InsertOrder :one
INSERT INTO orders (user_id, symbol, side, type, status, quantity, price, stop_price)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateOrderStatus :one
UPDATE orders
SET filled_quantity = $2, avg_fill_price = $3,
    status = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CancelOrder :one
UPDATE orders
SET status = 'canceled', updated_at = NOW()
WHERE id = $1 AND status = 'pending'
RETURNING *;