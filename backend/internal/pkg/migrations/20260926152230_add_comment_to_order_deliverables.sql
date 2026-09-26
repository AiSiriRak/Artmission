-- +goose Up
-- Add optional comment column to order_deliverables
ALTER TABLE order_deliverables
ADD COLUMN comment TEXT;

-- +goose Down
-- Remove comment column from order_deliverables
ALTER TABLE order_deliverables
DROP COLUMN comment;