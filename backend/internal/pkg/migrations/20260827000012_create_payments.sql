-- +goose Up
-- The escrow record for an order (EPIC 7): central hold, released or
-- refunded on completion/cancellation. backup.sql had no equivalent table
-- at all — its `Ordert.Status` conflated order progress with payment
-- state, which can't represent "paid but still in progress" separately
-- from "in progress, not yet paid".
-- TODO: This is just a draft, not finalized
CREATE TABLE IF NOT EXISTS payments (
    id       uuid PRIMARY KEY,
    order_id uuid NOT NULL UNIQUE REFERENCES orders (id) ON DELETE RESTRICT,
    -- Total escrowed for this order; equals orders.price at charge time.
    amount   numeric(12, 2) NOT NULL CHECK (amount >= 0),
    status   text NOT NULL DEFAULT 'PENDING'
             CHECK (status IN ('PENDING', 'HELD', 'SETTLED', 'FAILED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS payments;
