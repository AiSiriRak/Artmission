-- +goose Up
UPDATE artist_profiles AS ap
SET review_score = (
    SELECT ROUND(AVG(r.rating)::numeric, 1)
    FROM reviews AS r
    WHERE r.artist_id = ap.user_id
);

-- +goose Down
SELECT 1;
