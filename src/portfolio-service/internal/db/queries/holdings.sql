-- name: GetHoldingByAccountIdAndSymbol :one
SELECT * FROM holdings
WHERE account_id = $1 AND symbol = $2;

-- name: GetHoldingsByAccountId :many
SELECT * FROM holdings
WHERE account_id = $1;

-- name: InsertHolding :one
-- Used when buying for the FIRST time
INSERT INTO holdings ( account_id, symbol, quantity, avg_cost)
VALUES ( $1, $2, $3, $4)
RETURNING *;

-- name: UpdateHolding :one
-- Used when buying MORE or selling some
UPDATE holdings
SET quantity = $3,
    avg_cost = $4,
    updated_at = NOW()
WHERE account_id = $1 AND symbol = $2
RETURNING *;
