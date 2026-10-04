-- +goose Up
-- Add deleted_at column to artworks for soft delete
ALTER TABLE artworks
ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

-- Create index on deleted_at for fast query performance
CREATE INDEX IF NOT EXISTS idx_artworks_deleted_at ON artworks (deleted_at);

-- Orders now reference an artwork directly, so remove the descriptive snapshots.
ALTER TABLE orders
DROP CONSTRAINT orders_artwork_id_artist_id_fkey,
DROP COLUMN IF EXISTS artwork_name_snapshot,
DROP COLUMN IF EXISTS artwork_description_snapshot,
DROP COLUMN IF EXISTS minimum_deadline_days_snapshot;

-- Keep validating artwork ownership and prevent deleting referenced artworks.
-- artwork_id remains nullable for historical orders whose artwork was deleted
-- before this migration.
ALTER TABLE orders
ADD CONSTRAINT orders_artwork_id_artist_id_fkey
FOREIGN KEY (artwork_id, artist_id)
REFERENCES artworks (id, artist_id)
ON DELETE RESTRICT;

-- Rename price column
ALTER TABLE orders
RENAME COLUMN price_satang_snapshot TO price_satang_order;

-- +goose Down
-- Revert price column name
ALTER TABLE orders
RENAME COLUMN price_satang_order TO price_satang_snapshot;

-- Restore the nullable artwork reference and snapshot columns.
ALTER TABLE orders
DROP CONSTRAINT orders_artwork_id_artist_id_fkey,
ALTER COLUMN artwork_id DROP NOT NULL,
ADD COLUMN IF NOT EXISTS artwork_name_snapshot text,
ADD COLUMN IF NOT EXISTS artwork_description_snapshot text,
ADD COLUMN IF NOT EXISTS minimum_deadline_days_snapshot integer CHECK (minimum_deadline_days_snapshot > 0);

-- Set snapshot data according to artwork_id
UPDATE orders
SET
    artwork_name_snapshot = COALESCE(
        (SELECT artworks.name FROM artworks WHERE artworks.id = orders.artwork_id),
        orders.name
    ),
    artwork_description_snapshot = COALESCE(
        (SELECT artworks.description FROM artworks WHERE artworks.id = orders.artwork_id),
        ''
    ),
    minimum_deadline_days_snapshot = COALESCE(
        (SELECT artworks.minimum_deadline_days FROM artworks WHERE artworks.id = orders.artwork_id),
        1
    );

-- Revert snapshot columns to not null
ALTER TABLE orders
ALTER COLUMN artwork_name_snapshot
SET
    NOT NULL,
ALTER COLUMN artwork_description_snapshot
SET
    NOT NULL,
ALTER COLUMN minimum_deadline_days_snapshot
SET
    NOT NULL;

-- Restore the original behavior: deleting an artwork clears the order reference.
ALTER TABLE orders
ADD CONSTRAINT orders_artwork_id_artist_id_fkey
FOREIGN KEY (artwork_id, artist_id)
REFERENCES artworks (id, artist_id)
ON DELETE SET NULL (artwork_id);

-- Drop index
DROP INDEX IF EXISTS idx_artworks_deleted_at;

-- Remove deleted_at from artworks
ALTER TABLE artworks
DROP COLUMN IF EXISTS deleted_at;