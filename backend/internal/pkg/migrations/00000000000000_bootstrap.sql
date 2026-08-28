-- Placeholder migration so the goose/embed wiring has at least one file to
-- embed and apply. Domain schema migrations land alongside the modules that
-- need them.

-- +goose Up
SELECT 1;

-- +goose Down
SELECT 1;
