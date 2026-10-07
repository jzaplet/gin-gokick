-- +goose Up
CREATE TABLE auth_failures (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    scope text NOT NULL,
    network cidr NOT NULL,
    failed_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX auth_failures_scope_network_failed_at ON auth_failures (scope, network, failed_at);

CREATE INDEX auth_failures_scope_failed_at ON auth_failures (scope, failed_at);

-- +goose Down
DROP TABLE auth_failures;
