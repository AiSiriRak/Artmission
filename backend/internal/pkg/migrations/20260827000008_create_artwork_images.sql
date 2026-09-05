-- +goose Up
CREATE TABLE IF NOT EXISTS artwork_images (
    id                 uuid PRIMARY KEY,
    artwork_id         uuid NOT NULL REFERENCES artworks (id) ON DELETE CASCADE,
    original_image_url text NOT NULL,
    preview_image_url  text NOT NULL,
    sort_order         integer NOT NULL CHECK (sort_order >= 0),
    created_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT artwork_images_artwork_id_sort_order_key UNIQUE (artwork_id, sort_order)
);

-- +goose Down
DROP TABLE IF EXISTS artwork_images;
