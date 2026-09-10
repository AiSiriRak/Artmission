-- +goose Up
ALTER TABLE artist_profiles
    ADD COLUMN profile_image_key text;

-- +goose Down
ALTER TABLE artist_profiles
    DROP COLUMN profile_image_key;
