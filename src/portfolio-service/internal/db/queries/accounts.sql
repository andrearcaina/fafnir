-- name: GetAccountByUserId :one
SELECT * FROM accounts WHERE user_id = $1;

-- name: InsertAccount :one
INSERT INTO accounts (user_id, account_number, account_type, currency, balance)
VALUES ( $1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateAccountBalance :one
UPDATE accounts
SET
    balance = balance + $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1;
