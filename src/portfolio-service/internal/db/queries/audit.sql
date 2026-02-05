-- name: InsertAuditLog :one
INSERT INTO transactions ( account_id, transaction_type, amount, description, reference_id
) VALUES ( $1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTransactionsByAccountId :many
SELECT * FROM transactions
WHERE account_id = $1
ORDER BY created_at DESC;
