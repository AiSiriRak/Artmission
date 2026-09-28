-- +goose Up
-- Add optional comment column to order_deliverables
ALTER TABLE order_deliverables
ADD COLUMN comment TEXT;

-- Create constraint to check that comment will only available when decision is REJECTED
ALTER TABLE order_deliverables ADD CONSTRAINT chk_decision_comment_rule CHECK (
    (
        decision = 'REJECTED'
        AND comment IS NOT NULL
        AND trim(comment) <> ''
    )
    OR (
        decision <> 'REJECTED'
        AND comment IS NULL
    )
);

-- +goose Down
-- Remove constraint chk_decision_comment_rule from order_deliverables
ALTER TABLE order_deliverables
DROP CONSTRAINT IF EXISTS chk_decision_comment_rule;

-- Remove comment column from order_deliverables
ALTER TABLE order_deliverables
DROP COLUMN comment;