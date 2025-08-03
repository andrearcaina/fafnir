-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- stores user information
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- stores role information
CREATE TABLE roles (
    name VARCHAR(50) PRIMARY KEY,
    description TEXT
);

-- stores permission information
CREATE TABLE permissions (
    name VARCHAR(50) PRIMARY KEY,
    description TEXT
);

-- links roles to permissions
CREATE TABLE roles_permissions (
    role_name VARCHAR(50) REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

-- links users to roles
CREATE TABLE users_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_name VARCHAR(50) REFERENCES roles(name) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_name)
);

-- +goose Down
DROP TABLE IF EXISTS users_roles;
DROP TABLE IF EXISTS roles_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS users;