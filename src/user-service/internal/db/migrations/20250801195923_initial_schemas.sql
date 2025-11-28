-- +goose Up
CREATE TABLE user_profiles (
   id UUID PRIMARY KEY, -- this gets set by the auth service when a user registers and logs in (the JWT will contain this ID)
   first_name VARCHAR(255) NOT NULL,
   last_name VARCHAR(255) NOT NULL,
   email VARCHAR(255) UNIQUE NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE user_profiles;
