-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id            uuid PRIMARY KEY,
    username      text NOT NULL,
    email         text NOT NULL,
    password_hash text NOT NULL,
    role          text NOT NULL CHECK (role IN ('customer', 'artist', 'admin')),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    -- Soft delete: orders/reviews/reports/messages FK to users and must
    -- survive account deletion for history/audit.
    deleted_at    timestamptz
);

-- Unique only among live users so a deleted email can be reused.
CREATE UNIQUE INDEX users_email_key ON users (email) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS users;
