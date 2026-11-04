BEGIN;

CREATE TABLE wallets (
    id BIGSERIAL PRIMARY KEY,
    currency VARCHAR(16) NOT NULL,
    character_id INTEGER REFERENCES characters(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT wallets_owner_check CHECK (
        (currency = 'credits' AND character_id IS NOT NULL AND user_id IS NULL) OR
        (currency = 'cartel_coins' AND user_id IS NOT NULL AND character_id IS NULL)
    ),
    UNIQUE (character_id, currency),
    UNIQUE (user_id, currency)
);

CREATE TRIGGER wallets_updated_at BEFORE UPDATE ON wallets
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();

CREATE TABLE wallet_entries (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    source VARCHAR(32) COLLATE "C" NOT NULL CHECK (source ~ '^[a-z][a-z0-9_]{0,31}$'),
    operation_key VARCHAR(128) COLLATE "C" NOT NULL CHECK (operation_key ~ '^[A-Za-z0-9][A-Za-z0-9:._/-]{0,127}$'),
    amount BIGINT NOT NULL CHECK (amount <> 0 AND amount <> '-9223372036854775808'::BIGINT),
    balance_before BIGINT NOT NULL CHECK (balance_before >= 0),
    balance_after BIGINT NOT NULL CHECK (balance_after >= 0),
    description VARCHAR(200) NOT NULL CHECK (char_length(btrim(description)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (wallet_id, source, operation_key),
    CHECK (balance_after::NUMERIC = balance_before::NUMERIC + amount::NUMERIC)
);
CREATE INDEX wallet_entries_history_idx ON wallet_entries (wallet_id, id DESC);

-- History is append-only while its owner exists. Normal owner deletion still
-- cascades, matching the lifecycle of characters and accounts elsewhere.
CREATE FUNCTION protect_wallet_entry() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'DELETE' AND NOT EXISTS (SELECT 1 FROM wallets WHERE id = OLD.wallet_id) THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION 'wallet: entries are immutable' USING ERRCODE = '23514';
END $$;
CREATE TRIGGER wallet_entries_immutable BEFORE UPDATE OR DELETE ON wallet_entries
    FOR EACH ROW EXECUTE FUNCTION protect_wallet_entry();

CREATE FUNCTION protect_wallet_owner() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        IF (NEW.id, NEW.currency, NEW.character_id, NEW.user_id)
            IS DISTINCT FROM (OLD.id, OLD.currency, OLD.character_id, OLD.user_id) THEN
            RAISE EXCEPTION 'wallet: owner and currency are immutable' USING ERRCODE = '23514';
        END IF;
        RETURN NEW;
    END IF;
    IF (OLD.character_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM characters WHERE id = OLD.character_id))
        OR (OLD.user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM users WHERE id = OLD.user_id)) THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION 'wallet: delete the owner instead of its wallet' USING ERRCODE = '23514';
END $$;
CREATE TRIGGER wallets_owner_immutable BEFORE UPDATE OR DELETE ON wallets
    FOR EACH ROW EXECUTE FUNCTION protect_wallet_owner();

CREATE FUNCTION create_owner_wallet() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF TG_TABLE_NAME = 'users' THEN
        INSERT INTO wallets(currency, user_id) VALUES ('cartel_coins', NEW.id);
    ELSE
        INSERT INTO wallets(currency, character_id) VALUES ('credits', NEW.id);
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER users_create_wallet AFTER INSERT ON users
    FOR EACH ROW EXECUTE FUNCTION create_owner_wallet();
CREATE TRIGGER characters_create_wallet AFTER INSERT ON characters
    FOR EACH ROW EXECUTE FUNCTION create_owner_wallet();

INSERT INTO wallets(currency, user_id) SELECT 'cartel_coins', id FROM users;
INSERT INTO wallets(currency, character_id) SELECT 'credits', id FROM characters;

COMMIT;
