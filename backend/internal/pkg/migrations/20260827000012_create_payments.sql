-- +goose Up
-- One escrow summary per order; money movements are recorded separately
-- in payment_transactions.
CREATE TABLE IF NOT EXISTS payments (
    id            uuid PRIMARY KEY,
    order_id      uuid NOT NULL UNIQUE REFERENCES orders (id) ON DELETE RESTRICT,
    amount_satang bigint NOT NULL CHECK (amount_satang >= 0),
    status        text NOT NULL DEFAULT 'PENDING'
                  CHECK (status IN ('PENDING', 'HELD', 'SETTLED', 'FAILED')),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS payments;
