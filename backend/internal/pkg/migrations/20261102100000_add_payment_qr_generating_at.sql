-- +goose Up
-- Short-lived claim that serializes QR generation per payment without holding a pooled connection.
ALTER TABLE payments
    ADD COLUMN qr_generating_at timestamptz;

-- +goose Down
ALTER TABLE payments
    DROP COLUMN IF EXISTS qr_generating_at;
