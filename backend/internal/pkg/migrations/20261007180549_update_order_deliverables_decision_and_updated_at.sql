-- +goose Up
-- Add updated_at column
ALTER TABLE order_deliverables
ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT current_timestamp;

-- Drop old check constraint and update decision column
ALTER TABLE order_deliverables
DROP CONSTRAINT IF EXISTS order_deliverables_decision_check;

-- Pending deliverables are now stored as WAIT instead of NULL.
DROP INDEX IF EXISTS order_deliverables_order_id_pending_key;

-- Set old decision from NULL to 'WAIT'
UPDATE order_deliverables
SET
    decision = 'WAIT'
WHERE
    decision IS NULL;

-- Add constraint to validate ('WAIT', 'APPROVED', 'REJECTED') and set NOT NULL
ALTER TABLE order_deliverables
ALTER COLUMN decision SET DEFAULT 'WAIT',
ALTER COLUMN decision SET NOT NULL,
ADD CONSTRAINT order_deliverables_decision_check 
CHECK (decision IN ('WAIT', 'APPROVED', 'REJECTED'));

-- Add pending_key
CREATE UNIQUE INDEX order_deliverables_order_id_pending_key
    ON order_deliverables (order_id) WHERE decision = 'WAIT';

-- +goose Down
-- Delete constraint decision
ALTER TABLE order_deliverables
DROP CONSTRAINT IF EXISTS order_deliverables_decision_check;

-- Delete pending_key
DROP INDEX IF EXISTS order_deliverables_order_id_pending_key;

-- Revert WAIT back to NULL
ALTER TABLE order_deliverables
ALTER COLUMN decision DROP DEFAULT,
ALTER COLUMN decision DROP NOT NULL;

UPDATE order_deliverables 
SET decision = NULL 
WHERE decision = 'WAIT';

ALTER TABLE order_deliverables
ADD CONSTRAINT order_deliverables_decision_check 
CHECK (decision IN ('APPROVED', 'REJECTED'));

CREATE UNIQUE INDEX order_deliverables_order_id_pending_key
    ON order_deliverables (order_id) WHERE decision IS NULL;

-- Delete updated_at
ALTER TABLE order_deliverables
DROP COLUMN IF EXISTS updated_at;