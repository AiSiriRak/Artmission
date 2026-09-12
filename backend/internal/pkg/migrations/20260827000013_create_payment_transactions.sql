-- +goose Up
-- Append-only ledger of charges, artist releases, and customer refunds
-- against an order's payment.
CREATE TABLE IF NOT EXISTS payment_transactions (
    id            uuid PRIMARY KEY,
    payment_id    uuid NOT NULL REFERENCES payments (id) ON DELETE CASCADE,
    type          text NOT NULL
                  CHECK (type IN ('CHARGE', 'RELEASE_TO_ARTIST', 'REFUND_TO_CUSTOMER')),
    amount_satang bigint NOT NULL CHECK (amount_satang >= 0),
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX payment_transactions_payment_id_idx ON payment_transactions (payment_id);

-- +goose Down
DROP TABLE IF EXISTS payment_transactions;
