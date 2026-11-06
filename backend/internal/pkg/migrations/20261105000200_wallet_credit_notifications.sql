-- +goose Up
ALTER TABLE notifications DROP CONSTRAINT notifications_type_check;
ALTER TABLE notifications ADD CONSTRAINT notifications_type_check CHECK (type IN (
    'GENERAL', 'ORDER_CREATED', 'ORDER_ACCEPTED', 'ORDER_DECLINED',
    'ORDER_CANCELED', 'PAYMENT_SECURED', 'DRAFT_SUBMITTED',
    'REVISION_REQUESTED', 'ORDER_COMPLETED', 'FINAL_ARTWORK_READY',
    'DEADLINE_EXTENSION_REQUESTED', 'DEADLINE_EXTENSION_APPROVED',
    'DEADLINE_EXTENSION_REJECTED', 'WALLET_CREDITED', 'REFUND_CREDITED'
));

-- +goose Down
-- Refuse rollback while wallet cards exist; do not delete or reinterpret their history.
-- +goose StatementBegin
DO $$
BEGIN
    LOCK TABLE notifications IN ACCESS EXCLUSIVE MODE;
    IF EXISTS (SELECT 1 FROM notifications WHERE type IN ('WALLET_CREDITED', 'REFUND_CREDITED')) THEN
        RAISE EXCEPTION 'Cannot roll back wallet notification types while wallet notifications exist';
    END IF;
END;
$$;
-- +goose StatementEnd
ALTER TABLE notifications DROP CONSTRAINT notifications_type_check;
ALTER TABLE notifications ADD CONSTRAINT notifications_type_check CHECK (type IN (
    'GENERAL', 'ORDER_CREATED', 'ORDER_ACCEPTED', 'ORDER_DECLINED',
    'ORDER_CANCELED', 'PAYMENT_SECURED', 'DRAFT_SUBMITTED',
    'REVISION_REQUESTED', 'ORDER_COMPLETED', 'FINAL_ARTWORK_READY',
    'DEADLINE_EXTENSION_REQUESTED', 'DEADLINE_EXTENSION_APPROVED',
    'DEADLINE_EXTENSION_REJECTED'
));
