-- +goose Up
CREATE INDEX idx_orders_active_deadline ON orders (deadline_at)
WHERE status IN ('PENDING', 'NOT_PAID', 'IN_PROCESS');
CREATE INDEX idx_orders_pending_created ON orders (created_at)
WHERE status = 'PENDING';
-- +goose Down
DROP INDEX idx_orders_pending_created;
DROP INDEX idx_orders_active_deadline;