package db_test

import (
	"os"
	"path/filepath"
	"testing"
)

const walletMigration = "20261104000400_wallets"

func TestWalletMigrationBackfillAndRollback(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, walletMigration)
	snapshot := `SELECT jsonb_build_array((SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM characters t),(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM users t),(SELECT jsonb_agg(to_jsonb(t) ORDER BY id) FROM locations t))::text`
	var before, after string
	if err := s.pool.QueryRow(ctx, snapshot).Scan(&before); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		executeMigration(t, ctx, s, walletMigration+".up.sql")
		var wallets, zero int
		if err := s.pool.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE balance=0) FROM wallets`).Scan(&wallets, &zero); err != nil {
			t.Fatal(err)
		}
		if wallets != 5 || zero != 5 {
			t.Fatalf("backfill wallets=%d zero=%d", wallets, zero)
		}
		executeMigration(t, ctx, s, walletMigration+".down.sql")
		if err := s.pool.QueryRow(ctx, snapshot).Scan(&after); err != nil || before != after {
			t.Fatalf("rollback changed owners/world: %v", err)
		}
	}
	executeMigration(t, ctx, s, walletMigration+".up.sql")
	mustExec(t, ctx, s, `INSERT INTO users(email,login,password) VALUES('walletnew@example.test','walletnew','unused');
 INSERT INTO characters(user_id,faction_id,class_id,name) SELECT 3,faction_id,id,'New wallet owner' FROM classes WHERE code='sith_warrior'`)
	var wallets int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM wallets WHERE balance=0`).Scan(&wallets); err != nil || wallets != 7 {
		t.Fatalf("new owner wallets=%d err=%v", wallets, err)
	}
}

func TestWalletConstraintsAndOwnerCascade(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	for _, query := range []string{
		`INSERT INTO wallets(currency,user_id) VALUES('credits',1)`,
		`INSERT INTO wallets(currency,character_id) VALUES('cartel_coins',1)`,
		`INSERT INTO wallets(currency,character_id,user_id) VALUES('credits',1,1)`,
		`INSERT INTO wallets(currency) VALUES('credits')`,
		`INSERT INTO wallets(currency,user_id) VALUES('unknown',1)`,
		`INSERT INTO wallets(currency,character_id) VALUES('credits',1)`,
		`INSERT INTO wallets(currency,user_id) VALUES('cartel_coins',1)`,
		`INSERT INTO wallets(currency,character_id) VALUES('credits',99999)`,
		`UPDATE wallets SET balance=-1 WHERE character_id=1`,
		`UPDATE wallets SET character_id=3 WHERE character_id=1`,
		`DELETE FROM wallets WHERE character_id=1`,
	} {
		if _, err := s.pool.Exec(ctx, query); err == nil {
			t.Fatalf("accepted invalid wallet change: %s", query)
		}
	}
	mustExec(t, ctx, s, `INSERT INTO wallet_entries(wallet_id,source,operation_key,amount,balance_before,balance_after,description)
 SELECT id,'test','first',20,0,20,'Начисление' FROM wallets WHERE character_id=1 OR user_id=1`)
	for _, query := range []string{
		`UPDATE wallet_entries SET amount=10,balance_after=10`,
		`DELETE FROM wallet_entries`,
		`INSERT INTO wallet_entries(wallet_id,source,operation_key,amount,balance_before,balance_after,description) SELECT id,'test','first',20,0,20,'Дубль' FROM wallets WHERE character_id=1`,
		`INSERT INTO wallet_entries(wallet_id,source,operation_key,amount,balance_before,balance_after,description) SELECT id,'test','zero',0,0,0,'Ноль' FROM wallets WHERE character_id=1`,
		`INSERT INTO wallet_entries(wallet_id,source,operation_key,amount,balance_before,balance_after,description) SELECT id,'test','bad',20,0,19,'Ошибка' FROM wallets WHERE character_id=1`,
	} {
		if _, err := s.pool.Exec(ctx, query); err == nil {
			t.Fatalf("accepted invalid ledger change: %s", query)
		}
	}
	mustExec(t, ctx, s, `DELETE FROM characters WHERE id=1`)
	var wallets, entries int
	if err := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM wallets),(SELECT count(*) FROM wallet_entries)`).Scan(&wallets, &entries); err != nil || wallets != 4 || entries != 1 {
		t.Fatalf("character cascade: %d/%d %v", wallets, entries, err)
	}
	mustExec(t, ctx, s, `DELETE FROM users WHERE id=1`)
	if err := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM wallets),(SELECT count(*) FROM wallet_entries)`).Scan(&wallets, &entries); err != nil || wallets != 2 || entries != 0 {
		t.Fatalf("account cascade: %d/%d %v", wallets, entries, err)
	}
}

func TestWalletRollbackRefusesHistoryAndFunds(t *testing.T) {
	for _, setup := range []string{
		`UPDATE wallets SET balance=1 WHERE character_id=1`,
		`INSERT INTO wallet_entries(wallet_id,source,operation_key,amount,balance_before,balance_after,description) SELECT id,'test','recorded',1,0,1,'Награда' FROM wallets WHERE character_id=1`,
	} {
		t.Run(setup[:6], func(t *testing.T) {
			s, ctx := characterTestStorage(t, true)
			mustExec(t, ctx, s, setup)
			sql, err := os.ReadFile(filepath.Join("..", "..", "migrations", walletMigration+".down.sql"))
			if err != nil {
				t.Fatal(err)
			}
			conn, err := s.pool.Acquire(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Release()
			if _, err = conn.Exec(ctx, string(sql)); err == nil {
				t.Fatal("rollback discarded money/history")
			}
			if _, err = conn.Exec(ctx, "ROLLBACK"); err != nil {
				t.Fatal(err)
			}
			var count int
			if err = conn.QueryRow(ctx, `SELECT count(*) FROM wallets`).Scan(&count); err != nil || count != 5 {
				t.Fatalf("failed rollback lost wallets: %d %v", count, err)
			}
		})
	}
}
