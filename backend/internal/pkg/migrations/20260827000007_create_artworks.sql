-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS artworks (
    id                    uuid PRIMARY KEY,
    artist_id             uuid NOT NULL REFERENCES artist_profiles (user_id) ON DELETE CASCADE,
    category_id           uuid NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    name                  text NOT NULL,
    description           text NOT NULL,
    price_satang          bigint NOT NULL CHECK (price_satang >= 0),
    minimum_deadline_days integer NOT NULL CHECK (minimum_deadline_days > 0),
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    -- Required by orders' composite foreign key, which also verifies ownership.
    CONSTRAINT artworks_id_artist_id_key UNIQUE (id, artist_id)
);

CREATE INDEX artworks_search_trgm_idx
    ON artworks USING gin ((name || ' ' || description) gin_trgm_ops);
CREATE INDEX artworks_artist_id_idx ON artworks (artist_id);
CREATE INDEX artworks_category_id_idx ON artworks (category_id);
CREATE INDEX artworks_price_satang_idx ON artworks (price_satang);
CREATE INDEX users_username_trgm_idx ON users USING gin (username gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS users_username_trgm_idx;
DROP TABLE IF EXISTS artworks;
DROP EXTENSION IF EXISTS pg_trgm;
