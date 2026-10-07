-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    email text NOT NULL UNIQUE CHECK (email = lower(email)),
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    locale text NOT NULL CHECK (locale ~ '^[a-z]{2}_[A-Z]{2}$')
);

-- +goose Down
DROP TABLE users;
