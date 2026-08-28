-- +goose Up
CREATE TABLE orders (
    id           uuid PRIMARY KEY,
    customer_id  uuid NOT NULL REFERENCES users (id),
    artist_id    uuid NOT NULL REFERENCES users (id),
    description  text NOT NULL,
    category     text,
    style        text,
    price        numeric(12, 2),
    status       text NOT NULL DEFAULT 'PENDING'
                 CHECK (status IN ('PENDING', 'IN_PROGRESS', 'COMPLETE', 'CANCELLED')),
    deadline     timestamptz,
    completed_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX orders_customer_id_idx ON orders (customer_id);
CREATE INDEX orders_artist_id_idx ON orders (artist_id);
CREATE INDEX orders_status_idx ON orders (status);

-- +goose Down
DROP TABLE orders;
