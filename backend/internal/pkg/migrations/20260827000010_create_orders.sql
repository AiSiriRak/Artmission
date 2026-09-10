-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    id                             uuid PRIMARY KEY,
    customer_id                    uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    artist_id                      uuid NOT NULL REFERENCES artist_profiles (user_id) ON DELETE RESTRICT,
    artwork_id                     uuid,

    -- Immutable purchase terms retained when the source artwork changes or is deleted.
    artwork_name_snapshot          text NOT NULL,
    artwork_description_snapshot   text NOT NULL,
    price_satang_snapshot          bigint NOT NULL CHECK (price_satang_snapshot >= 0),
    minimum_deadline_days_snapshot integer NOT NULL CHECK (minimum_deadline_days_snapshot > 0),
    preview_image_url_snapshot     text NOT NULL,

    customer_description           text NOT NULL,
    -- Duration is the original agreement; deadline_at is set when work starts and may move after an approved extension.
    selected_deadline_days         integer NOT NULL
                                   CHECK (selected_deadline_days >= minimum_deadline_days_snapshot),
    deadline_at                    timestamptz,
    status                         text NOT NULL DEFAULT 'PENDING'
                                   CHECK (status IN ('PENDING', 'NOT_PAID', 'IN_PROCESS', 'SUCCESS', 'CANCEL')),


    completed_at                   timestamptz,
    created_at                     timestamptz NOT NULL DEFAULT now(),
    updated_at                     timestamptz NOT NULL DEFAULT now(),

    -- Required by reviews' composite foreign key, which verifies both order parties.
    CONSTRAINT orders_id_customer_id_artist_id_key UNIQUE (id, customer_id, artist_id),
    -- Verifies the artwork belongs to artist_id; deletion clears only artwork_id so snapshots survive.
    CONSTRAINT orders_artwork_id_artist_id_fkey
        FOREIGN KEY (artwork_id, artist_id)
        REFERENCES artworks (id, artist_id)
        ON DELETE SET NULL (artwork_id)
);

CREATE INDEX orders_customer_id_idx ON orders (customer_id);
CREATE INDEX orders_artist_id_idx ON orders (artist_id);
CREATE INDEX orders_status_idx ON orders (status);
CREATE INDEX orders_artwork_id_idx ON orders (artwork_id);

-- +goose Down
DROP TABLE IF EXISTS orders;
