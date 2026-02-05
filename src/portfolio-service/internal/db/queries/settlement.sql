-- name: UpdateAccountBalance :one
UPDATE accounts
SET 
    balance = balance + $2, 
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpsertHolding :one
INSERT INTO holdings (account_id, symbol, quantity, avg_cost, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
ON CONFLICT (account_id, symbol)
DO UPDATE SET
    quantity = holdings.quantity + EXCLUDED.quantity,
    avg_cost = CASE
        WHEN holdings.quantity + EXCLUDED.quantity = 0 THEN 0
        ELSE (holdings.quantity * holdings.avg_cost + EXCLUDED.quantity * EXCLUDED.avg_cost) / (holdings.quantity + EXCLUDED.quantity)
    END,
    updated_at = NOW()
RETURNING *;

-- name: DecreaseHolding :one
UPDATE holdings
SET quantity = quantity - $3, updated_at = NOW()
WHERE account_id = $1 AND symbol = $2
RETURNING *;
