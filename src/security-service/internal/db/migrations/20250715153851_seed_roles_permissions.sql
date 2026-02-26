-- +goose Up
INSERT INTO roles_permissions (role_name, permission_name) VALUES
    ('member', 'view_stocks'),
    ('member', 'order_stocks'),
    ('member', 'manage_own_portfolio'),
    ('member', 'manage_own_watchlist'),
    ('member', 'manage_own_profile'),
    ('member', 'manage_own_accounts'),
    ('admin', 'view_admin_dashboard'),
    ('admin', 'manage_own_profile'),
    ('admin', 'view_audit_logs'),
    ('admin', 'view_users'),
    ('admin', 'deactivate_users'),
    ('admin', 'reactivate_users'),
    ('admin', 'manage_roles');

-- +goose Down
DELETE FROM roles_permissions
WHERE (role_name = 'member' AND permission_name IN (
    'view_stocks',
    'order_stocks',
    'manage_portfolio',
    'manage_watchlist',
    'manage_profile',
    'manage_own_accounts'
))
OR (role_name = 'admin' AND permission_name IN (
    'view_users',
    'view_admin_dashboard',
    'manage_own_profile',
    'view_audit_logs',
    'deactivate_users',
    'reactivate_users',
    'manage_roles'
));
