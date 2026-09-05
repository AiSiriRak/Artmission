-- +goose Up
-- Final images delivered for an order; unlike artwork sample images, these
-- are immutable order history and are visible to the customer after success.
CREATE TABLE IF NOT EXISTS order_deliverables (
    id                 uuid PRIMARY KEY,
    order_id           uuid NOT NULL REFERENCES orders (id) ON DELETE RESTRICT,
    original_image_url text NOT NULL,
    preview_image_url  text NOT NULL,
    sort_order         integer NOT NULL CHECK (sort_order >= 0),
    created_at         timestamptz NOT NULL DEFAULT now(),
    UNIQUE (order_id, sort_order)
);

-- +goose Down
DROP TABLE IF EXISTS order_deliverables;
