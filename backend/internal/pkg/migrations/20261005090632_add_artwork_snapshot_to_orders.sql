-- +goose Up
-- 1. Add column artwork_snapshot to orders
ALTER TABLE orders
ADD COLUMN IF NOT EXISTS artwork_snapshot jsonb NOT NULL DEFAULT '{}'::jsonb;

-- 2. Backfill artwork snapshot data from artworks and artwork_styles
UPDATE orders
SET artwork_snapshot = jsonb_build_object(
    'artwork_name', artworks.name,
    'category_id', artworks.category_id,
    'style_ids', COALESCE(
        (
            SELECT jsonb_agg(artwork_styles.style_id)
            FROM artwork_styles
            WHERE artwork_styles.artwork_id = artworks.id
        ),
        '[]'::jsonb
    )
)
FROM artworks
WHERE orders.artwork_id = artworks.id;

-- 3. Allow NULL on artwork_id in orders
ALTER TABLE orders
ALTER COLUMN artwork_id DROP NOT NULL;

-- 4. Change Foreign Key constraint to ON DELETE SET NULL (artwork_id)
ALTER TABLE orders
DROP CONSTRAINT IF EXISTS orders_artwork_id_artist_id_fkey;

ALTER TABLE orders
ADD CONSTRAINT orders_artwork_id_artist_id_fkey
FOREIGN KEY (artwork_id, artist_id)
REFERENCES artworks (id, artist_id)
ON DELETE SET NULL (artwork_id);

-- 5. Remove soft delete mechanism from artworks
DROP INDEX IF EXISTS idx_artworks_deleted_at;

ALTER TABLE artworks
DROP COLUMN IF EXISTS deleted_at;


-- +goose Down
-- 1. Restore soft delete column and index on artworks
ALTER TABLE artworks
ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_artworks_deleted_at ON artworks (deleted_at);

-- 2. Revert Foreign Key constraint back to RESTRICT
ALTER TABLE orders
DROP CONSTRAINT IF EXISTS orders_artwork_id_artist_id_fkey;

ALTER TABLE orders
ADD CONSTRAINT orders_artwork_id_artist_id_fkey
FOREIGN KEY (artwork_id, artist_id)
REFERENCES artworks (id, artist_id)
ON DELETE RESTRICT;

-- 3. Drop artwork_snapshot column from orders
ALTER TABLE orders
DROP COLUMN IF EXISTS artwork_snapshot;