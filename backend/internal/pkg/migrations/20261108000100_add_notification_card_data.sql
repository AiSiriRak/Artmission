-- +goose Up
ALTER TABLE notifications
    ADD COLUMN order_id uuid REFERENCES orders (id) ON DELETE SET NULL,
    ADD COLUMN deliverable_id uuid REFERENCES order_deliverables (id) ON DELETE SET NULL,
    ADD COLUMN delay_request_id uuid REFERENCES order_delay_requests (id) ON DELETE SET NULL,
    DROP CONSTRAINT notifications_type_check;

-- Recover links from legacy pending/accepted notifications when an order and recipient match.
WITH legacy AS (
    SELECT
        id,
        substring(message FROM 'Order: ([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})[.] Status: (PENDING|NOT_PAID)[.]$')::uuid AS parsed_order_id,
        CASE
            WHEN title LIKE 'New Order:%' THEN 'ORDER_CREATED'
            WHEN title LIKE 'Order Accepted:%' THEN 'ORDER_ACCEPTED'
        END AS parsed_event
    FROM notifications
    WHERE type = 'ORDER_UPDATE'
      AND (title LIKE 'New Order:%' OR title LIKE 'Order Accepted:%')
)
UPDATE notifications AS n
SET
    type = legacy.parsed_event,
    order_id = o.id
FROM legacy
JOIN orders AS o ON o.id = legacy.parsed_order_id
WHERE n.id = legacy.id
  AND n.user_id IN (o.customer_id, o.artist_id);


UPDATE notifications SET type = 'GENERAL' WHERE type IN ('ORDER_UPDATE', 'REVIEW', 'SYSTEM');

ALTER TABLE notifications
    ALTER COLUMN type SET DEFAULT 'GENERAL',
    ADD CONSTRAINT notifications_type_check CHECK (type IN (
        'GENERAL', 'ORDER_CREATED', 'ORDER_ACCEPTED', 'ORDER_DECLINED',
        'ORDER_CANCELED', 'PAYMENT_SECURED', 'DRAFT_SUBMITTED',
        'REVISION_REQUESTED', 'ORDER_COMPLETED', 'FINAL_ARTWORK_READY',
        'DEADLINE_EXTENSION_REQUESTED', 'DEADLINE_EXTENSION_APPROVED',
        'DEADLINE_EXTENSION_REJECTED'
    ));

-- +goose Down
-- Refuse rollback rather than discard event types or references unsupported by the original table.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM notifications
        WHERE type <> 'GENERAL' OR order_id IS NOT NULL
           OR deliverable_id IS NOT NULL OR delay_request_id IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Cannot roll back notifications while workflow events or references exist';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE notifications DROP CONSTRAINT notifications_type_check;
UPDATE notifications SET type = 'SYSTEM';
ALTER TABLE notifications
    ALTER COLUMN type DROP DEFAULT,
    ADD CONSTRAINT notifications_type_check CHECK (type IN ('ORDER_UPDATE', 'REVIEW', 'SYSTEM')),
    DROP COLUMN delay_request_id,
    DROP COLUMN deliverable_id,
    DROP COLUMN order_id;
