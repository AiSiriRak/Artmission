-- +goose Up
-- 1:1 extension of users
CREATE TABLE IF NOT EXISTS bank_accounts (
    user_id        uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    bank_name      text NOT NULL,
    account_number text NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS bank_accounts;
