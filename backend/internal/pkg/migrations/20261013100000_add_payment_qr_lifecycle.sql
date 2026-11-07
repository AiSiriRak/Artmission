-- +goose Up
ALTER TABLE payments
    ADD COLUMN qr_payload text,
    ADD COLUMN qr_expire timestamptz,
    ADD COLUMN provider_transaction_ref text;

ALTER TABLE payments
    DROP CONSTRAINT payments_status_check,
    ADD CONSTRAINT payments_status_check
        CHECK (status IN ('PENDING', 'HELD', 'SETTLED', 'FAILED', 'EXPIRED'));

CREATE UNIQUE INDEX payments_provider_transaction_ref_key
    ON payments (provider_transaction_ref)
    WHERE provider_transaction_ref IS NOT NULL;

-- +goose StatementBegin
DO $$
DECLARE
    cutover_at timestamptz := clock_timestamp();
BEGIN
    UPDATE payments AS p
    SET qr_expire = cutover_at + INTERVAL '24 hours',
        updated_at = cutover_at
    FROM orders AS o
    WHERE o.id = p.order_id
      AND o.status = 'NOT_PAID'
      AND p.status = 'PENDING';

    INSERT INTO payments (
        id,
        order_id,
        amount_satang,
        status,
        qr_expire,
        created_at,
        updated_at
    )
    SELECT
        gen_random_uuid(),
        o.id,
        o.price_satang_order,
        'PENDING',
        cutover_at + INTERVAL '24 hours',
        cutover_at,
        cutover_at
    FROM orders AS o
    WHERE o.status = 'NOT_PAID'
      AND NOT EXISTS (
          SELECT 1
          FROM payments AS p
          WHERE p.order_id = o.id
      );
END;
$$;
-- +goose StatementEnd

-- +goose Down
DROP INDEX IF EXISTS payments_provider_transaction_ref_key;

ALTER TABLE payments
    DROP CONSTRAINT payments_status_check,
    ADD CONSTRAINT payments_status_check
        CHECK (status IN ('PENDING', 'HELD', 'SETTLED', 'FAILED')),
    DROP COLUMN provider_transaction_ref,
    DROP COLUMN qr_expire,
    DROP COLUMN qr_payload;
