-- +goose Up
CREATE TABLE IF NOT EXISTS artist_styles (
    artist_id uuid NOT NULL REFERENCES artist_profiles (user_id) ON DELETE CASCADE,
    style_id  uuid NOT NULL REFERENCES styles (id) ON DELETE RESTRICT,
    PRIMARY KEY (artist_id, style_id)
);

CREATE INDEX IF NOT EXISTS artist_styles_style_id_artist_id_idx
    ON artist_styles (style_id, artist_id);

-- +goose Down
DROP TABLE IF EXISTS artist_styles;
