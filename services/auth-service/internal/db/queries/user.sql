-- name: RegisterUser :one
INSERT INTO users (id, email, password_hash, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
RETURNING id, email;

-- name: UpdatePassword :exec
UPDATE users
SET password_hash = $1, updated_at = NOW()
WHERE id = $2;

-- name: GetUserByEmail :one
SELECT id, email, password_hash
FROM users
WHERE email = $1;

-- name: GetUserById :one
SELECT id, email, password_hash
FROM users
WHERE id = $1;