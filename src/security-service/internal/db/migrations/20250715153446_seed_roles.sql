-- +goose Up
INSERT INTO roles (name, description) VALUES
  ('member', 'Standard user with default access to app'),
  ('admin', 'Admin with more permissions');

-- +goose Down
DELETE FROM roles WHERE name IN ('member', 'admin');
