-- +goose Up
ALTER TABLE artwork_images
    ADD COLUMN image_url text;

-- Portfolio previews were public before this API existed, so retain the preview URL.
UPDATE artwork_images
SET image_url = preview_image_key;

ALTER TABLE artwork_images
    ALTER COLUMN image_url SET NOT NULL,
    DROP COLUMN original_image_key,
    DROP COLUMN preview_image_key;

-- +goose Down
ALTER TABLE artwork_images
    ADD COLUMN original_image_key text,
    ADD COLUMN preview_image_key text;

UPDATE artwork_images
SET original_image_key = image_url,
    preview_image_key = image_url;

ALTER TABLE artwork_images
    ALTER COLUMN original_image_key SET NOT NULL,
    ALTER COLUMN preview_image_key SET NOT NULL,
    DROP COLUMN image_url;
