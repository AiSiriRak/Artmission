-- +goose Up
DROP INDEX IF EXISTS users_username_key;

-- +goose Down
CREATE UNIQUE INDEX users_username_key ON users (username) WHERE deleted_at IS NULL;
