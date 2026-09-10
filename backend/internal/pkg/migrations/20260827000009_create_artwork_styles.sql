-- +goose Up
CREATE TABLE IF NOT EXISTS artwork_styles (
    artwork_id uuid NOT NULL REFERENCES artworks (id) ON DELETE CASCADE,
    style_id   uuid NOT NULL REFERENCES styles (id) ON DELETE RESTRICT,
    PRIMARY KEY (artwork_id, style_id)
);

CREATE INDEX artwork_styles_style_id_artwork_id_idx ON artwork_styles (style_id, artwork_id);

-- +goose Down
DROP TABLE IF EXISTS artwork_styles;
