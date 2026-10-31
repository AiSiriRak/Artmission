-- +goose Up
ALTER TABLE notifications
    ADD COLUMN event_type text NOT NULL DEFAULT 'GENERAL'
        CHECK (event_type IN ('GENERAL', 'ORDER_CREATED', 'ORDER_ACCEPTED')),
    ADD COLUMN order_id uuid REFERENCES orders (id) ON DELETE SET NULL,
    ADD COLUMN order_status text
        CHECK (order_status IN ('PENDING', 'NOT_PAID')),
    ADD COLUMN deadline_at timestamptz;

-- Recover links from legacy pending/accepted notifications when an order and recipient match.
WITH legacy AS (
    SELECT
        id,
        substring(message FROM 'Order: ([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})[.] Status: (PENDING|NOT_PAID)[.]$')::uuid AS parsed_order_id,
        substring(message FROM 'Status: (PENDING|NOT_PAID)[.]$') AS parsed_status,
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
    event_type = legacy.parsed_event,
    order_id = o.id,
    order_status = legacy.parsed_status,
    deadline_at = o.deadline_at
FROM legacy
JOIN orders AS o ON o.id = legacy.parsed_order_id
WHERE n.id = legacy.id
  AND n.user_id IN (o.customer_id, o.artist_id);

-- +goose Down
ALTER TABLE notifications
    DROP COLUMN deadline_at,
    DROP COLUMN order_status,
    DROP COLUMN order_id,
    DROP COLUMN event_type;
