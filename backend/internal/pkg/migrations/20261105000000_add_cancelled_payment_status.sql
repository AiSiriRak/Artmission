-- +goose Up
ALTER TABLE payments DROP CONSTRAINT payments_status_check,
    ADD CONSTRAINT payments_status_check
        CHECK (status IN ('PENDING', 'HELD', 'SETTLED', 'FAILED', 'EXPIRED', 'CANCELLED'));

-- +goose Down
-- Never reinterpret cancelled payments or delete history merely to permit rollback.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM payments WHERE status = 'CANCELLED') THEN
        RAISE EXCEPTION 'Cannot roll back CANCELLED payment status while cancelled payments exist';
    END IF;
END;
$$;
-- +goose StatementEnd
ALTER TABLE payments DROP CONSTRAINT payments_status_check,
    ADD CONSTRAINT payments_status_check
        CHECK (status IN ('PENDING', 'HELD', 'SETTLED', 'FAILED', 'EXPIRED'));
