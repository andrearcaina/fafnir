-- this is for seeding
-- name: InsertUserRoleWithID :one
INSERT INTO users_roles (user_id, role_name)
VALUES ($1, $2)
RETURNING user_id, role_name;

-- name: CheckUserPermission :one
SELECT EXISTS (
    SELECT 1
    FROM users_roles ur
    JOIN roles_permissions rp ON ur.role_name = rp.role_name
    WHERE ur.user_id = $1 AND rp.permission_name = $2
) AS has_permission;
