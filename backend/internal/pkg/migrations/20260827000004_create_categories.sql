-- +goose Up
-- Admin-managed art categories.
CREATE TABLE IF NOT EXISTS categories (
    id         uuid PRIMARY KEY,
    label      text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS categories;
