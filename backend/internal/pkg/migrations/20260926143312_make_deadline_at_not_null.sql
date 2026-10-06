-- +goose Up
-- Fix existing NULL values with current timestamp to prevent constraint violation
UPDATE orders
SET
    deadline_at = NOW ()
WHERE
    deadline_at IS NULL;

-- Apply NOT NULL constraint to deadline_at
ALTER TABLE orders
ALTER COLUMN deadline_at
SET
    NOT NULL;

-- +goose Down
-- Remove NOT NULL constraint from deadline_at
ALTER TABLE orders
ALTER COLUMN deadline_at
DROP NOT NULL;