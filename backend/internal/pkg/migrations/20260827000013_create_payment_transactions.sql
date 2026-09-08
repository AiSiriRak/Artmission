-- +goose Up
-- Append-only settlement ledger, one row per money movement against a
-- payment. Reshaped from backup.sql's `Transaction` table (GET_PAID/
-- PAY_TO tied straight to an order) into movements against the escrow
-- record instead, because customer-initiated cancellation (US6.3) splits
-- a single payment 50/50 between a refund and a payout — a single
-- status column can't represent a split, but two ledger rows can.
-- TODO: This is just a draft, not finalized
CREATE TABLE IF NOT EXISTS payment_transactions (
    id         uuid PRIMARY KEY,
    payment_id uuid NOT NULL REFERENCES payments (id) ON DELETE CASCADE,
    -- CHARGE: customer -> escrow (US7.2).
    -- RELEASE_TO_ARTIST: escrow -> artist, full (US7.3) or half of a
    --   customer-cancel split (US6.3).
    -- REFUND_TO_CUSTOMER: escrow -> customer, full (US6.2/US6.4) or half
    --   of a customer-cancel split (US6.3).
    type       text NOT NULL
               CHECK (type IN ('CHARGE', 'RELEASE_TO_ARTIST', 'REFUND_TO_CUSTOMER')),
    amount     numeric(12, 2) NOT NULL CHECK (amount >= 0),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX payment_transactions_payment_id_idx ON payment_transactions (payment_id);

-- +goose Down
DROP TABLE IF EXISTS payment_transactions;
