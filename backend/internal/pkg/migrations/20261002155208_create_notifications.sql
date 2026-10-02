-- +goose Up
CREATE TABLE
    notifications (
        id uuid PRIMARY KEY,
        user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
        title TEXT NOT NULL,
        message TEXT NOT NULL,
        type TEXT NOT NULL CHECK (type IN ('ORDER_UPDATE', 'REVIEW', 'SYSTEM')),
        is_read BOOLEAN NOT NULL DEFAULT FALSE,
        created_at timestamptz NOT NULL DEFAULT now ()
    );

CREATE INDEX idx_notifications_user_unread ON notifications (user_id)
WHERE
    is_read = FALSE;

CREATE INDEX idx_notifications_user_feed ON notifications (user_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_notifications_user_feed;

DROP INDEX IF EXISTS idx_notifications_user_unread;

DROP TABLE IF EXISTS notifications;