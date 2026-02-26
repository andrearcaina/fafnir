-- +goose Up
INSERT INTO permissions (name, description) VALUES
    ('view_stocks', 'View stock metadata and quotes'),
    ('order_stocks', 'Order stocks (buy/sell)'),
    ('manage_own_portfolio', 'Manage assets in portfolio (crud on portfolio table)'),
    ('manage_own_watchlist', 'Add or remove from watchlist'),
    ('manage_own_profile', 'Updates their own data in user profile (crud on users table)'),
    ('manage_own_accounts', 'Manage user accounts (crud on accounts table)'),
    ('view_users', 'View all user accounts (for admins)'),
    ('view_admin_dashboard', 'View admin dashboard (for admins)'),
    ('view_audit_logs', 'View audit logs (for admins)'),
    ('deactivate_users', 'Deactivate user accounts (for admins)'),
    ('reactivate_users', 'Reactivate user accounts (for admins)'),
    ('manage_roles', 'Manage user roles (for admins)');
-- +goose Down
DELETE FROM permissions WHERE name IN (
    'view_stocks',
    'order_stocks',
    'manage_own_portfolio',
    'manage_own_watchlist',
    'manage_own_profile',
    'manage_own_accounts',
    'view_users',
    'view_admin_dashboard',
    'view_audit_logs',
    'deactivate_users',
    'reactivate_users',
    'manage_roles'
);
