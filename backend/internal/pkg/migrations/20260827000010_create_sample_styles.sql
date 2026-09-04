-- +goose Up
-- Which styles a given sample depicts; a sample may depict more than one.
CREATE TABLE IF NOT EXISTS sample_styles (
    sample_id uuid NOT NULL REFERENCES artist_samples (id) ON DELETE CASCADE,
    style_id  uuid NOT NULL REFERENCES styles (id) ON DELETE CASCADE,
    PRIMARY KEY (sample_id, style_id)
);

-- Used when searching for samples which depict a specific style.
CREATE INDEX sample_styles_style_id_idx ON sample_styles (style_id);

-- +goose Down
DROP TABLE IF EXISTS sample_styles;
