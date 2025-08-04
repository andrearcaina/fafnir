-- +goose NO TRANSACTION

-- +goose Up
CREATE DATABASE auth_db;
CREATE DATABASE user_db;
CREATE DATABASE security_db;

-- +goose Down
DROP DATABASE auth_db;
DROP DATABASE user_db;
DROP DATABASE security_db;