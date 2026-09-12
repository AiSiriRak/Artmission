-- +goose Up
CREATE TABLE IF NOT EXISTS order_deliverables (
    id                 uuid PRIMARY KEY,
    order_id           uuid NOT NULL REFERENCES orders (id) ON DELETE RESTRICT,
    version            integer NOT NULL CHECK (version >= 1), -- The ceiling enforcement is handled by application instead

    -- NULL: awaiting the customer's decision.
    -- APPROVED: order becomes SUCCESS.
    -- REJECTED: artist may submit the next version.
    decision           text CHECK (decision IN ('APPROVED', 'REJECTED')),

    -- These images are private, only accessible via presigned URLs
    original_image_key text NOT NULL,
    preview_image_key  text NOT NULL,

    created_at         timestamptz NOT NULL DEFAULT now(),
    UNIQUE (order_id, version)
);

-- At most one undecided draft per order: blocks the artist from submitting
-- a new version while the customer hasn't approved/rejected the current one.
CREATE UNIQUE INDEX order_deliverables_order_id_pending_key
    ON order_deliverables (order_id) WHERE decision IS NULL;

-- At most one approved deliverable per order: exactly one final artwork,
-- even though up to 3 versions may have been attempted.
CREATE UNIQUE INDEX order_deliverables_order_id_approved_key
    ON order_deliverables (order_id) WHERE decision = 'APPROVED';

-- +goose Down
DROP TABLE IF EXISTS order_deliverables;
