-- this is for seeding
-- name: InsertUserRoleWithID :one
INSERT INTO users_roles (user_id, role_name)
VALUES ($1, $2)
RETURNING user_id, role_name;