-- +goose Up
-- TODO: This is just a draft, not finalized
CREATE TABLE IF NOT EXISTS orders (
    id          uuid PRIMARY KEY,
    customer_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    artist_id   uuid NOT NULL REFERENCES artist_profiles (user_id) ON DELETE RESTRICT,
    description text NOT NULL,
    category_id uuid REFERENCES categories (id) ON DELETE SET NULL,
    style_id    uuid REFERENCES styles (id) ON DELETE SET NULL,
    price       numeric(12, 2) NOT NULL CHECK (price >= 0),
    status      text NOT NULL DEFAULT 'PENDING'
                CHECK (status IN ('PENDING', 'IN_PROGRESS', 'COMPLETE', 'CANCELLED')),
    -- The delivered file (US6.5), set once together when status ->
    -- COMPLETE. Not its own table: it's 1:1 with this order, created
    -- exactly once, never revised, and has no independent existence
    -- (unlike artist_samples, it's never browsed, reordered, or
    -- edited) — see artist_samples.sql for the artist-managed portfolio
    -- side. US2.5 promotes one into a sample by copying these two URLs
    -- (plus description/category_id/style_id) into a new artist_samples
    -- row.
    delivered_original_url text,
    delivered_preview_url  text,
    deadline     timestamptz,
    completed_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX orders_customer_id_idx ON orders (customer_id);
CREATE INDEX orders_artist_id_idx ON orders (artist_id);
CREATE INDEX orders_status_idx ON orders (status);

-- +goose Down
DROP TABLE IF EXISTS orders;
