-- +goose Up
-- Admin-managed reference art styles.
CREATE TABLE IF NOT EXISTS styles (
    id         uuid PRIMARY KEY,
    label      text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS styles;
