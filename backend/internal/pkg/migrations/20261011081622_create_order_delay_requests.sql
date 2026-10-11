-- +goose Up
CREATE TABLE
    order_delay_requests (
        id                      uuid        PRIMARY KEY,
        order_id                uuid        NOT NULL    REFERENCES orders (id) ON DELETE RESTRICT,
        original_deadline_at    timestamptz NOT NULL,
        deadline_request_at     timestamptz NOT NULL,
        reason                  text        NOT NULL    DEFAULT 'OTHER' CHECK (
            reason IN ('NEED_MORE_TIME', 'WAITING_FOR_FEEDBACK', 'PERSONAL_REASON', 'HIGH_WORKLOAD', 'TECHNICAL_ISSUE', 'OTHER')
        ),
        explanation             text        NOT NULL,
        customer_decision       text        NOT NULL    DEFAULT 'WAIT' CHECK (
            customer_decision IN ('WAIT', 'APPROVED', 'REJECTED')
        ),
        created_at              timestamptz NOT NULL    DEFAULT now (),
        updated_at              timestamptz NOT NULL    DEFAULT now (),

        CONSTRAINT chk_deadline_request_after_original CHECK (deadline_request_at > original_deadline_at)
    );

CREATE UNIQUE INDEX idx_pending_order_delay_req 
ON order_delay_requests (order_id)
WHERE customer_decision = 'WAIT';

-- +goose Down
DROP INDEX IF EXISTS idx_pending_order_delay_req;
DROP TABLE IF EXISTS order_delay_requests;