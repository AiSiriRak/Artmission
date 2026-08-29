-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id            uuid PRIMARY KEY,
    username      text NOT NULL UNIQUE,
    email         text NOT NULL UNIQUE,
    first_name    text NOT NULL,
    last_name     text NOT NULL,
    phone_number  text NOT NULL,
    password_hash text NOT NULL,
    role          text NOT NULL CHECK (role IN ('customer', 'artist', 'admin')),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    -- Soft delete because orders/reviews/reports/messages FK to users and must survive account deletion for history/audit
    deleted_at    timestamptz
);

-- +goose Down
DROP TABLE IF EXISTS users;
