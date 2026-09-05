-- +goose Up
-- 1:1 extension of users where role = 'artist'.
-- Enforced by application constraint, not DB constraint.
CREATE TABLE IF NOT EXISTS artist_profiles (
    user_id      uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    description  text,
    -- Denormalized: caches the average rating so it doesn't need to be
    -- recomputed on every page view. Maintained by application code on
    -- each review write; eventually consistent, which is acceptable since
    -- it's a display-only value (unlike, e.g., payment amounts).
    review_score numeric(2, 1),
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX artist_profiles_review_score_idx
    ON artist_profiles (review_score) WHERE review_score IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS artist_profiles;
