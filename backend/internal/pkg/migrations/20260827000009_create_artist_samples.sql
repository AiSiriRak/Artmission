-- +goose Up
-- The artist's portfolio: fully artist-managed (create/edit/delete/reorder,
-- US2.3/US2.4), unlike orders' delivered file, which is an immutable
-- historical fact owned by the order (see orders.delivered_original_url).
-- Promoting a delivered order into a sample (US2.5) copies its fields
-- into a new row here — it never shares a row with the order, so deleting
-- a sample can never touch order history.
-- TODO: This is just a draft, not finalized
CREATE TABLE IF NOT EXISTS artist_samples (
    id                 uuid PRIMARY KEY,
    artist_id          uuid NOT NULL REFERENCES artist_profiles (user_id) ON DELETE CASCADE,
    category_id        uuid REFERENCES categories (id) ON DELETE SET NULL,
    name               text NOT NULL,
    description        text,
    price              numeric(10, 2) CHECK (price >= 0),
    original_image_url text NOT NULL,
    preview_image_url  text NOT NULL,
    -- Artist-controlled display position on the profile (US2.3 says
    -- samples are "shown on the profile"; created_at can't represent the
    -- artist manually reordering them).
    -- TODO: not decided yet — do we actually need manual reordering, or is
    -- newest-first/some other fixed order enough? Nice to have either way,
    -- revisit before building the reorder endpoint.
    sort_order         integer NOT NULL DEFAULT 0,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX artist_samples_artist_id_idx ON artist_samples (artist_id);
CREATE INDEX artist_samples_category_id_idx ON artist_samples (category_id);

-- +goose Down
DROP TABLE IF EXISTS artist_samples;
