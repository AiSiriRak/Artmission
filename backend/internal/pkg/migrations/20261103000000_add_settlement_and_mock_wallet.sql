-- +goose Up
ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS source text NOT NULL DEFAULT 'EXTERNAL' CHECK (source IN ('EXTERNAL', 'MOCK_WALLET')),
    ADD COLUMN IF NOT EXISTS settled_at timestamptz,
    ADD COLUMN IF NOT EXISTS settlement_outcome text CHECK (settlement_outcome IN ('COMPLETED', 'DEADLINE_EXPIRED', 'ARTIST_CANCELLED', 'CUSTOMER_CANCELLED'));

ALTER TABLE payment_transactions
    ADD COLUMN IF NOT EXISTS transfer_status text CHECK (transfer_status IN ('PENDING', 'SUCCESS', 'FAILED')),
    ADD COLUMN IF NOT EXISTS transfer_ref text,
    ADD COLUMN IF NOT EXISTS transferred_at timestamptz,
    ADD COLUMN IF NOT EXISTS failure_reason text;

-- Existing transfer rows are allocations awaiting execution, never evidence of a transfer.
UPDATE payment_transactions SET transfer_status = 'PENDING'
WHERE type <> 'CHARGE' AND transfer_status IS NULL;

-- The legacy settlement migration used the same columns but weaker metadata checks.
-- Transfer metadata checks are added after the wallet migration fills recipient IDs.
-- +goose StatementBegin
DO $$
DECLARE
    legacy_settlement boolean := COALESCE((
        SELECT is_applied FROM goose_db_version WHERE version_id = 20261101100000 ORDER BY id DESC LIMIT 1
    ), false);
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = 'payments'::regclass AND conname = 'payments_settlement_metadata_check') THEN
        ALTER TABLE payments ADD CONSTRAINT payments_settlement_metadata_check CHECK (
            (settled_at IS NULL AND settlement_outcome IS NULL) OR
            (status = 'SETTLED' AND settled_at IS NOT NULL AND settlement_outcome IS NOT NULL)
        ) NOT VALID;
    END IF;
    IF NOT legacy_settlement THEN
        ALTER TABLE payments VALIDATE CONSTRAINT payments_settlement_metadata_check;
    END IF;

    -- Reuse the legacy (payment_id,type) unique index even though its name differs.
    IF NOT EXISTS (
        SELECT 1 FROM pg_index i
        WHERE i.indrelid = 'payment_transactions'::regclass AND i.indisunique AND i.indisvalid
          AND i.indpred IS NULL AND i.indexprs IS NULL AND i.indnatts = 2
          AND i.indkey[0] = (SELECT attnum FROM pg_attribute WHERE attrelid = i.indrelid AND attname = 'payment_id')
          AND i.indkey[1] = (SELECT attnum FROM pg_attribute WHERE attrelid = i.indrelid AND attname = 'type')
    ) THEN
        -- Deliberately fail on duplicate history; never silently delete financial records.
        CREATE UNIQUE INDEX payment_transactions_payment_type_key ON payment_transactions (payment_id, type);
    END IF;
END;
$$;
-- +goose StatementEnd

CREATE TABLE wallet_entries (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    request_id uuid,
    payment_transaction_id uuid UNIQUE REFERENCES payment_transactions(id) ON DELETE RESTRICT,
    kind text NOT NULL CHECK (kind IN ('TOP_UP', 'CHARGE', 'RELEASE_TO_ARTIST', 'REFUND_TO_CUSTOMER')),
    amount_satang bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, request_id),
    CHECK ((kind = 'TOP_UP' AND amount_satang > 0 AND request_id IS NOT NULL AND payment_transaction_id IS NULL) OR
           (kind = 'CHARGE' AND amount_satang <= 0 AND request_id IS NULL AND payment_transaction_id IS NOT NULL) OR
           (kind IN ('RELEASE_TO_ARTIST', 'REFUND_TO_CUSTOMER') AND amount_satang > 0 AND request_id IS NULL AND payment_transaction_id IS NOT NULL))
);
CREATE INDEX wallet_entries_user_created_idx ON wallet_entries(user_id, created_at DESC, id DESC);

-- +goose Down
-- Lock before inspecting so a credit cannot commit between the guard and DROP.
-- +goose StatementBegin
DO $$
BEGIN
    LOCK TABLE wallet_entries IN ACCESS EXCLUSIVE MODE;
    IF EXISTS (SELECT 1 FROM wallet_entries) THEN
        RAISE EXCEPTION 'Cannot roll back wallet ledger migration while wallet entries exist';
    END IF;
END;
$$;
-- +goose StatementEnd
DROP TABLE wallet_entries;
DROP INDEX IF EXISTS payment_transactions_payment_type_key;
ALTER TABLE payment_transactions DROP CONSTRAINT IF EXISTS payment_transactions_transfer_check;
ALTER TABLE payments DROP CONSTRAINT payments_settlement_metadata_check, DROP COLUMN source;

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT COALESCE((SELECT is_applied FROM goose_db_version WHERE version_id = 20261101100000 ORDER BY id DESC LIMIT 1), false) THEN
        ALTER TABLE payment_transactions DROP COLUMN transfer_status, DROP COLUMN transfer_ref,
            DROP COLUMN transferred_at, DROP COLUMN failure_reason;
        ALTER TABLE payments DROP COLUMN settled_at, DROP COLUMN settlement_outcome;
    END IF;
END;
$$;
-- +goose StatementEnd
