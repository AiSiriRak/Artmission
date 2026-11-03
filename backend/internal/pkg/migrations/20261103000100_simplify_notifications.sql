-- +goose Up
ALTER TABLE notifications
    ADD COLUMN deliverable_id uuid REFERENCES order_deliverables (id) ON DELETE SET NULL,
    ADD COLUMN delay_request_id uuid REFERENCES order_delay_requests (id) ON DELETE SET NULL,
    DROP CONSTRAINT notifications_type_check;

UPDATE notifications SET type = event_type;

ALTER TABLE notifications
    ALTER COLUMN type SET DEFAULT 'GENERAL',
    ADD CONSTRAINT notifications_type_check CHECK (type IN (
        'GENERAL', 'ORDER_CREATED', 'ORDER_ACCEPTED', 'ORDER_DECLINED',
        'ORDER_CANCELED', 'PAYMENT_SECURED', 'DRAFT_SUBMITTED',
        'REVISION_REQUESTED', 'ORDER_COMPLETED', 'FINAL_ARTWORK_READY',
        'DEADLINE_EXTENSION_REQUESTED', 'DEADLINE_EXTENSION_APPROVED',
        'DEADLINE_EXTENSION_REJECTED'
    )),
    DROP COLUMN event_type,
    DROP COLUMN order_status,
    DROP COLUMN deadline_at;

-- +goose Down
-- Refuse rollback rather than discard events or references unsupported by the previous schema.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM notifications
        WHERE type NOT IN ('GENERAL', 'ORDER_CREATED', 'ORDER_ACCEPTED')
           OR deliverable_id IS NOT NULL OR delay_request_id IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Cannot roll back notifications while newer events or workflow references exist';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE notifications
    ADD COLUMN event_type text NOT NULL DEFAULT 'GENERAL'
        CHECK (event_type IN ('GENERAL', 'ORDER_CREATED', 'ORDER_ACCEPTED')),
    ADD COLUMN order_status text CHECK (order_status IN ('PENDING', 'NOT_PAID')),
    ADD COLUMN deadline_at timestamptz;

UPDATE notifications SET event_type = type,
    order_status = CASE type
        WHEN 'ORDER_CREATED' THEN 'PENDING'
        WHEN 'ORDER_ACCEPTED' THEN 'NOT_PAID'
        ELSE NULL END;
UPDATE notifications AS n SET deadline_at = o.deadline_at
FROM orders AS o WHERE o.id = n.order_id;

ALTER TABLE notifications DROP CONSTRAINT notifications_type_check;
UPDATE notifications SET type = CASE WHEN type = 'GENERAL' THEN 'SYSTEM' ELSE 'ORDER_UPDATE' END;
ALTER TABLE notifications
    ALTER COLUMN type DROP DEFAULT,
    ADD CONSTRAINT notifications_type_check CHECK (type IN ('ORDER_UPDATE', 'REVIEW', 'SYSTEM')),
    DROP COLUMN deliverable_id,
    DROP COLUMN delay_request_id;
