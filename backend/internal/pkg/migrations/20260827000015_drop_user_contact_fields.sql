-- +goose Up
ALTER TABLE users DROP COLUMN first_name;
ALTER TABLE users DROP COLUMN last_name;
ALTER TABLE users DROP COLUMN phone_number;

-- +goose Down
ALTER TABLE users ADD COLUMN first_name text NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN last_name text NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN phone_number text NOT NULL DEFAULT '';
