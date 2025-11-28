-- name: InsertUserProfileById :one
INSERT INTO user_profiles (id, first_name, last_name, email, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id, first_name, last_name;

-- name: GetUserProfileById :one
SELECT id, first_name, last_name
FROM user_profiles
WHERE id = $1;

-- name: DeleteUserProfileById :exec
DELETE FROM user_profiles
WHERE id = $1;
