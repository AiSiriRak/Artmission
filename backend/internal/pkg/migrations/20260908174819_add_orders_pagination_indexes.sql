-- +goose Up
-- Composite index lets Postgres satisfy ViewOrders' default-sort query
-- (ORDER BY updated_at LIMIT/OFFSET, scoped by customer_id/artist_id)
-- with an ordered index scan instead of a full-table sort.
CREATE INDEX orders_customer_id_updated_at_id_idx ON orders (customer_id, updated_at, id);
CREATE INDEX orders_artist_id_updated_at_id_idx ON orders (artist_id, updated_at, id);

-- +goose Down
DROP INDEX IF EXISTS orders_customer_id_updated_at_id_idx;
DROP INDEX IF EXISTS orders_artist_id_updated_at_id_idx;
