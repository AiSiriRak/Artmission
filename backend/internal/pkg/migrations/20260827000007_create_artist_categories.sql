-- +goose Up
-- Which categories an artist works in.
CREATE TABLE IF NOT EXISTS artist_categories (
    artist_id   uuid NOT NULL REFERENCES artist_profiles (user_id) ON DELETE CASCADE,
    category_id uuid NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    PRIMARY KEY (artist_id, category_id)
);

-- Used when searching for artist who works in specific category.
CREATE INDEX artist_categories_category_id_idx ON artist_categories (category_id);

-- +goose Down
DROP TABLE IF EXISTS artist_categories;
