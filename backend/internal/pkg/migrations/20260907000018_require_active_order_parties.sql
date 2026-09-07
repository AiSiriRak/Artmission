-- +goose Up
CREATE OR REPLACE FUNCTION require_active_order_parties()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM 1
    FROM users
    WHERE id = NEW.customer_id AND deleted_at IS NULL
    FOR KEY SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'customer account is unavailable' USING ERRCODE = '23514';
    END IF;

    PERFORM 1
    FROM users
    WHERE id = NEW.artist_id AND deleted_at IS NULL
    FOR KEY SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'artist account is unavailable' USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER orders_require_active_parties
BEFORE INSERT OR UPDATE OF customer_id, artist_id, status ON orders
FOR EACH ROW
EXECUTE FUNCTION require_active_order_parties();

-- +goose Down
DROP TRIGGER IF EXISTS orders_require_active_parties ON orders;
DROP FUNCTION IF EXISTS require_active_order_parties();
