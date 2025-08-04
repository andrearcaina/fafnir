-- +goose Up
INSERT INTO roles_permissions (role_name, permission_name) VALUES
    ('member', 'view_stocks'),
    ('member', 'manage_portfolio'),
    ('member', 'manage_watchlist'),
    ('member', 'manage_profile'),
    ('member', 'manage_own_accounts'),
    ('admin', 'view_admin_dashboard'),
    ('admin', 'impersonate_user'),
    ('admin', 'manage_signup_codes'),
    ('admin', 'manage_profile'),
    ('admin', 'view_audit_logs'),
    ('admin', 'view_users'),
    ('admin', 'deactivate_users'),
    ('admin', 'reactivate_users'),
    ('admin', 'manage_roles');

-- +goose Down
DELETE FROM roles_permissions
WHERE (role_name = 'member' AND permission_name IN (
    'view_stocks',
    'manage_portfolio',
    'manage_watchlist',
    'manage_profile',
    'manage_own_accounts'
))
OR (role_name = 'admin' AND permission_name IN (
    'view_stocks',
    'manage_portfolio',
    'manage_watchlist',
    'manage_profile',
    'manage_own_accounts',
    'view_users',
    'deactivate_users',
    'manage_roles'
));
