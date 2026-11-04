BEGIN;

LOCK TABLE wallets, wallet_entries IN ACCESS EXCLUSIVE MODE;
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM wallet_entries) OR EXISTS (SELECT 1 FROM wallets WHERE balance <> 0) THEN
        RAISE EXCEPTION 'wallet: cannot roll back balances or a nonempty ledger';
    END IF;
END $$;

DROP TRIGGER characters_create_wallet ON characters;
DROP TRIGGER users_create_wallet ON users;
DROP TABLE wallet_entries;
DROP TABLE wallets;
DROP FUNCTION create_owner_wallet();
DROP FUNCTION protect_wallet_entry();
DROP FUNCTION protect_wallet_owner();

COMMIT;
