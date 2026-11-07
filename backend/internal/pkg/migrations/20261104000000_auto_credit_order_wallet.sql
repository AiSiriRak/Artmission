-- +goose Up
ALTER TABLE payment_transactions ADD COLUMN recipient_user_id uuid REFERENCES users(id) ON DELETE RESTRICT;
UPDATE payment_transactions t SET recipient_user_id =
    CASE t.type WHEN 'RELEASE_TO_ARTIST' THEN o.artist_id ELSE o.customer_id END
FROM payments p JOIN orders o ON o.id = p.order_id
WHERE t.payment_id = p.id AND t.type IN ('RELEASE_TO_ARTIST','REFUND_TO_CUSTOMER');

-- Add stricter metadata checks only after updating legacy recipient IDs. NOT VALID
-- preserves incomplete old SUCCESS metadata while still checking every new write.
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = 'payment_transactions'::regclass AND conname = 'payment_transactions_transfer_check') THEN
        ALTER TABLE payment_transactions ADD CONSTRAINT payment_transactions_transfer_check CHECK (
            (type = 'CHARGE' AND transfer_status IS NULL AND transfer_ref IS NULL AND transferred_at IS NULL AND failure_reason IS NULL) OR
            (type <> 'CHARGE' AND transfer_status IS NOT NULL AND
                ((transfer_status = 'SUCCESS' AND transfer_ref IS NOT NULL AND transferred_at IS NOT NULL AND failure_reason IS NULL) OR
                 (transfer_status IN ('PENDING', 'FAILED') AND transfer_ref IS NULL AND transferred_at IS NULL)))
        ) NOT VALID;
    END IF;
    IF NOT COALESCE((SELECT is_applied FROM goose_db_version WHERE version_id = 20261101100000 ORDER BY id DESC LIMIT 1), false) THEN
        ALTER TABLE payment_transactions VALIDATE CONSTRAINT payment_transactions_transfer_check;
    END IF;
END;
$$;
-- +goose StatementEnd

ALTER TABLE payment_transactions ADD CONSTRAINT payment_transactions_recipient_check CHECK (
    (type = 'CHARGE' AND recipient_user_id IS NULL) OR
    (type IN ('RELEASE_TO_ARTIST','REFUND_TO_CUSTOMER') AND recipient_user_id IS NOT NULL)
);
CREATE INDEX payment_transactions_recipient_idx ON payment_transactions(recipient_user_id, created_at DESC)
    WHERE recipient_user_id IS NOT NULL;

CREATE TABLE wallets (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
    balance_satang bigint NOT NULL DEFAULT 0 CHECK (balance_satang >= 0),
    updated_at timestamptz NOT NULL DEFAULT now()
);
-- Preserve existing journal balances; pending allocations are not credited by this migration.
INSERT INTO wallets(user_id, balance_satang, updated_at)
SELECT user_id, SUM(amount_satang)::bigint, MAX(created_at) FROM wallet_entries GROUP BY user_id;

-- +goose Down
-- Guard here too: Goose commits each migration separately before reaching the ledger Down.
-- +goose StatementBegin
DO $$
BEGIN
    LOCK TABLE wallet_entries, wallets IN ACCESS EXCLUSIVE MODE;
    IF EXISTS (SELECT 1 FROM wallet_entries) OR EXISTS (SELECT 1 FROM wallets WHERE balance_satang <> 0) THEN
        RAISE EXCEPTION 'Cannot roll back wallet credit migration while wallet entries or non-zero balances exist';
    END IF;
END;
$$;
-- +goose StatementEnd
DROP TABLE wallets;
DROP INDEX payment_transactions_recipient_idx;
ALTER TABLE payment_transactions DROP CONSTRAINT IF EXISTS payment_transactions_transfer_check;
ALTER TABLE payment_transactions DROP CONSTRAINT payment_transactions_recipient_check, DROP COLUMN recipient_user_id;
