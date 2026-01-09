-- name: InsertAuditLog :one
INSERT INTO transactions ( account_id, transaction_type, amount, description, reference_id
) VALUES ( $1, $2, $3, $4, $5)
RETURNING *;
