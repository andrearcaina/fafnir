-- for seeding
-- name: InsertUserProfileById :one
INSERT INTO user_profiles (id, first_name, last_name, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, first_name, last_name;

-- name: GetUserProfileById :one
SELECT id, first_name, last_name
FROM user_profiles
WHERE id = $1;