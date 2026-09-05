-- +goose Up
-- Existing payout destinations cannot safely be assigned a holder name from
-- user-profile data, so legacy rows must be completed by their owners.
ALTER TABLE bank_accounts
    ADD COLUMN account_holder_name text NOT NULL DEFAULT '';

ALTER TABLE bank_accounts
    ALTER COLUMN account_holder_name DROP DEFAULT;

-- +goose Down
ALTER TABLE bank_accounts
    DROP COLUMN account_holder_name;
