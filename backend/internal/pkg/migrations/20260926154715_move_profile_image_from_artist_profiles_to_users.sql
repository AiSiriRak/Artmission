-- +goose Up
-- Add profile_image_key column to users
ALTER TABLE users
ADD COLUMN IF NOT EXISTS profile_image_key TEXT;

-- Preserve profile images that were stored on artist profiles
UPDATE users AS u
SET profile_image_key = ap.profile_image_key
FROM artist_profiles AS ap
WHERE ap.user_id = u.id
	AND ap.profile_image_key IS NOT NULL;

-- Delete profile_image_key column to artist_profiles
ALTER TABLE artist_profiles
DROP COLUMN IF EXISTS profile_image_key;

-- +goose Down
-- Add profile_image_key column to artist_profiles
ALTER TABLE artist_profiles
ADD COLUMN IF NOT EXISTS profile_image_key TEXT;

-- Preserve profile images that were stored on users
UPDATE artist_profiles AS ap
SET profile_image_key = u.profile_image_key
FROM users AS u
WHERE ap.user_id = u.id
	AND u.profile_image_key IS NOT NULL;

-- Delete profile_image_key column to users
ALTER TABLE users
DROP COLUMN IF EXISTS profile_image_key;