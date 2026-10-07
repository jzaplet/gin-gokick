-- +goose Up
CREATE TABLE sessions (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    user_id uuid NOT NULL REFERENCES users ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    user_agent text NOT NULL CHECK (length(user_agent) <= 512),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_ip inet
);

CREATE INDEX sessions_user_id ON sessions (user_id);

CREATE INDEX sessions_expires_at ON sessions (expires_at);

-- +goose Down
DROP TABLE sessions;
