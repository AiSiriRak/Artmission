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

-- +goose Down
ALTER TABLE artist_profiles
    DROP COLUMN min_price_satang,
    DROP COLUMN max_price_satang;
