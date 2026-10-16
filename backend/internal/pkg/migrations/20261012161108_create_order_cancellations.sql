-- +goose Up
CREATE TABLE
    order_cancellations (
        id                      uuid        PRIMARY KEY,
        order_id                uuid        NOT NULL    REFERENCES orders (id) ON DELETE RESTRICT,
        reason                  text        NOT NULL    DEFAULT 'OTHER' CHECK (
            reason IN (
                    'ARTIST_DID_NOT_CONFIRM_IN_TIME',
                    'NOT_SUBMITTED_WITHIN_DEADLINE',

                    'UNABLE_TO_COMPLETE',
                    'UNABLE_TO_MEET_DEADLINE',
                    'CUSTOMER_CHANGED_REQUIREMENTS',
                    'CUSTOMER_UNRESPONSIVE',
                    'PERSONAL_OR_TECHNICAL_ISSUE',

                    'PLACED_ORDER_BY_MISTAKE',
                    'NO_LONGER_NEEDED',
                    'ARTIST_UNABLE_TO_MEET_NEW_REQUIREMENTS',
                    'ARTIST_UNRESPONSIVE',
                    'TAKING_TOO_LONG',

                    'MUTUAL_AGREEMENT',
                    'OTHER')
        ),
        explanation             text        NOT NULL,
        created_at              timestamptz NOT NULL    DEFAULT now ()
    );

CREATE UNIQUE INDEX idx_pending_order_cancel 
ON order_cancellations (order_id);

-- +goose Down
DROP INDEX IF EXISTS idx_pending_order_cancel;
DROP TABLE IF EXISTS order_cancellations;