-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS players (
    id text PRIMARY KEY,
    login text UNIQUE NOT NULL,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS players;

-- +goose StatementEnd
