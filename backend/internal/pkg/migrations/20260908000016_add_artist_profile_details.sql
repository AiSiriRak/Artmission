-- +goose Up
ALTER TABLE artist_profiles
    ADD COLUMN min_price_satang bigint,
    ADD COLUMN max_price_satang bigint,
    ADD CONSTRAINT artist_profiles_price_range_check CHECK (
        (min_price_satang IS NULL AND max_price_satang IS NULL)
        OR (
            min_price_satang IS NOT NULL
            AND max_price_satang IS NOT NULL
            AND min_price_satang >= 0
            AND max_price_satang >= 0
            AND min_price_satang <= max_price_satang
        )
    );

CREATE TABLE artist_styles (
    artist_id uuid NOT NULL REFERENCES artist_profiles (user_id) ON DELETE CASCADE,
    style_id  uuid NOT NULL REFERENCES styles (id) ON DELETE RESTRICT,
    PRIMARY KEY (artist_id, style_id)
);

CREATE INDEX artist_styles_style_id_artist_id_idx ON artist_styles (style_id, artist_id);

-- +goose Down
DROP TABLE artist_styles;

ALTER TABLE artist_profiles
    DROP COLUMN min_price_satang,
    DROP COLUMN max_price_satang;
