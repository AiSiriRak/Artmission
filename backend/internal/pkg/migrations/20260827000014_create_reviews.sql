-- +goose Up
CREATE TABLE IF NOT EXISTS reviews (
    id          uuid PRIMARY KEY,
    order_id    uuid NOT NULL UNIQUE,
    customer_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    artist_id   uuid NOT NULL REFERENCES artist_profiles (user_id) ON DELETE RESTRICT,
    rating      smallint NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment     text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    -- Ensures the review's customer and artist are the same parties as its order.
    CONSTRAINT reviews_order_id_customer_id_artist_id_fkey
        FOREIGN KEY (order_id, customer_id, artist_id)
        REFERENCES orders (id, customer_id, artist_id)
        ON DELETE RESTRICT
);

CREATE INDEX reviews_artist_id_idx ON reviews (artist_id);
CREATE INDEX reviews_customer_id_idx ON reviews (customer_id);

-- +goose Down
DROP TABLE IF EXISTS reviews;
