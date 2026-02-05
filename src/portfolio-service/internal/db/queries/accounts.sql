-- name: GetAccountByUserId :many
SELECT * FROM accounts WHERE user_id = $1;

-- name: InsertAccount :one
INSERT INTO accounts (user_id, account_number, account_type, currency, balance)
VALUES ( $1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1;
