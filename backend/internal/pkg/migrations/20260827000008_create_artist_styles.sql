-- +goose Up
-- Which styles an artist specializes in.
CREATE TABLE IF NOT EXISTS artist_styles (
    artist_id uuid NOT NULL REFERENCES artist_profiles (user_id) ON DELETE CASCADE,
    style_id  uuid NOT NULL REFERENCES styles (id) ON DELETE CASCADE,
    PRIMARY KEY (artist_id, style_id)
);

-- Used when searching for artist who specializes in specific style.
CREATE INDEX artist_styles_style_id_idx ON artist_styles (style_id);

-- +goose Down
DROP TABLE IF EXISTS artist_styles;
