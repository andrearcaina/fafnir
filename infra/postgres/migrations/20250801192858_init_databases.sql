-- +goose NO TRANSACTION

-- +goose Up
CREATE DATABASE auth_db;
CREATE DATABASE user_db;

-- +goose Down
DROP DATABASE auth_db;
DROP DATABASE user_db;
