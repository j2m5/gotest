package db_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/auth"
	"gotest/internal/db"
)

const timestampMigration = "20261025000600_timestamp_consistency"

var timestampTables = []string{"users", "sessions", "email_verifications", "factions", "classes", "abilities", "characters"}

func insertTimestampCredentials(t *testing.T, ctx context.Context, s *testStorage) {
	t.Helper()
	mustExec(t, ctx, s, `INSERT INTO sessions(user_id,token,expires_at) VALUES(1,'time-session','2027-02-01 12:00:00+00');
		INSERT INTO email_verifications(user_id,token,expired_at) VALUES(2,'time-verification','2027-02-01 12:00:00+00')`)
}

func storageInTimeZone(t *testing.T, ctx context.Context, s *testStorage, zone string) *testStorage {
	t.Helper()
	config := s.pool.Config()
	config.ConnConfig.RuntimeParams["timezone"] = zone
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return &testStorage{AuthRepository: db.NewAuthRepository(pool), pool: pool}
}

func timestampSnapshot(t *testing.T, ctx context.Context, s *testStorage) string {
	t.Helper()
	var snapshot strings.Builder
	for _, table := range timestampTables {
		var rows string
		if err := s.pool.QueryRow(ctx, "SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY id),'[]'::jsonb)::text FROM "+table+" t").Scan(&rows); err != nil {
			t.Fatal(err)
		}
		snapshot.WriteString(table + rows)
	}
	return snapshot.String()
}

func TestTimestampMigrationPreservesLegacyDatesAndRollback(t *testing.T) {
	for _, zone := range []string{"UTC", "Europe/Moscow", "America/New_York"} {
		t.Run(zone, func(t *testing.T) {
			s, ctx := characterTestStorageBefore(t, timestampMigration)
			s = storageInTimeZone(t, ctx, s, zone)
			insertTimestampCredentials(t, ctx, s)
			mustExec(t, ctx, s, `INSERT INTO users(email,login,password,email_verified_at)
				VALUES('later@example.test','later','unused','2026-04-05 10:20:30.123456')`)
			for _, table := range timestampTables {
				mustExec(t, ctx, s, "UPDATE "+table+" SET created_at='2026-01-15 12:34:56.123456',updated_at='2026-07-15 12:34:56.654321'")
			}
			// The old verification writer stored the same UTC timestamp in both columns.
			mustExec(t, ctx, s, `UPDATE users SET email_verified_at='2026-04-05 10:20:30.123456',updated_at='2026-04-05 10:20:30.123456' WHERE id=1;
				UPDATE users SET email_verified_at=NULL WHERE id=2;
				UPDATE characters SET activated_at='2026-07-15 12:34:56.654321' WHERE id=1`)
			before := timestampSnapshot(t, ctx, s)
			location, err := time.LoadLocation(zone)
			if err != nil {
				t.Fatal(err)
			}
			created := time.Date(2026, 1, 15, 12, 34, 56, 123456000, location)
			updated := time.Date(2026, 7, 15, 12, 34, 56, 654321000, location)
			verified := time.Date(2026, 4, 5, 10, 20, 30, 123456000, time.UTC)
			for pass := 0; pass < 2; pass++ {
				executeMigration(t, ctx, s, timestampMigration+".up.sql")
				// Read through another connection timezone to check instants, not wall times.
				reader := storageInTimeZone(t, ctx, s, "Asia/Kathmandu")
				for _, table := range timestampTables {
					var gotCreated, gotUpdated time.Time
					if err := reader.pool.QueryRow(ctx, "SELECT created_at,updated_at FROM "+table+" WHERE id=1").Scan(&gotCreated, &gotUpdated); err != nil {
						t.Fatal(err)
					}
					wantUpdated := updated
					if table == "users" {
						wantUpdated = verified
					}
					if !gotCreated.Equal(created) || !gotUpdated.Equal(wantUpdated) {
						t.Fatalf("%s dates moved: created=%v updated=%v", table, gotCreated, gotUpdated)
					}
				}
				var gotVerified, laterUpdated, gotActivated, sessionExpiry, verificationExpiry time.Time
				var nullVerified, nullActivated *time.Time
				if err := reader.pool.QueryRow(ctx, `SELECT
					(SELECT email_verified_at FROM users WHERE id=1),
					(SELECT updated_at FROM users WHERE id=3),
					(SELECT activated_at FROM characters WHERE id=1),
					(SELECT email_verified_at FROM users WHERE id=2),
					(SELECT activated_at FROM characters WHERE id=2),
					(SELECT expires_at FROM sessions WHERE id=1),
					(SELECT expired_at FROM email_verifications WHERE id=1)`).Scan(
					&gotVerified, &laterUpdated, &gotActivated, &nullVerified, &nullActivated, &sessionExpiry, &verificationExpiry); err != nil {
					t.Fatal(err)
				}
				if !gotVerified.Equal(verified) || !laterUpdated.Equal(updated) || !gotActivated.Equal(updated) || nullVerified != nil || nullActivated != nil {
					t.Fatal("event timestamps, later account updates or nullable dates changed")
				}
				expiry := time.Date(2027, 2, 1, 12, 0, 0, 0, time.UTC)
				if !sessionExpiry.Equal(expiry) || !verificationExpiry.Equal(expiry) {
					t.Fatal("already absolute credential expiry moved")
				}
				executeMigration(t, ctx, s, timestampMigration+".down.sql")
				if after := timestampSnapshot(t, ctx, s); after != before {
					t.Fatal("up/down changed original rows or timestamps")
				}
				var triggers int
				if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM pg_trigger WHERE tgrelid IN
					(SELECT oid FROM pg_class WHERE relnamespace=current_schema()::regnamespace)
					AND tgname LIKE '%_set_updated_at'`).Scan(&triggers); err != nil || triggers != 0 {
					t.Fatalf("down left timestamp triggers: %d %v", triggers, err)
				}
			}
		})
	}
}

func TestTimestampDefaultsAndVerificationAcrossTimeZones(t *testing.T) {
	base, ctx := characterTestStorage(t, true)
	for _, zone := range []string{"UTC", "Europe/Moscow", "America/New_York"} {
		t.Run(zone, func(t *testing.T) {
			s := storageInTimeZone(t, ctx, base, zone)
			var before, after, created, updated time.Time
			if err := s.pool.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&before); err != nil {
				t.Fatal(err)
			}
			email := strings.ToLower(strings.ReplaceAll(zone, "/", "-")) + "@example.test"
			if err := s.CreateRegistration(ctx, auth.Registration{Email: email, Login: zone, PasswordHash: "unused", VerificationTokenHash: zone, VerificationExpiresAt: before.Add(time.Hour)}); err != nil {
				t.Fatal(err)
			}
			if err := s.pool.QueryRow(ctx, `SELECT created_at,updated_at,clock_timestamp() FROM users WHERE email=$1`, email).Scan(&created, &updated, &after); err != nil {
				t.Fatal(err)
			}
			if !created.Equal(updated) || created.Before(before) || created.After(after) {
				t.Fatalf("insert defaults depend on timezone: created=%v updated=%v bounds=[%v,%v]", created, updated, before, after)
			}
			eventTime := before.Add(-time.Minute).In(time.FixedZone("event offset", 5*3600+45*60))
			if err := s.ConsumeVerification(ctx, zone, eventTime); err != nil {
				t.Fatal(err)
			}
			user, err := s.FindCredentialsByEmail(ctx, email)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.pool.QueryRow(ctx, `SELECT updated_at,clock_timestamp() FROM users WHERE id=$1`, user.User.ID).Scan(&updated, &after); err != nil {
				t.Fatal(err)
			}
			if user.User.EmailVerifiedAt == nil || user.User.EmailVerifiedAt.Sub(eventTime).Abs() > time.Microsecond || updated.Before(before) || updated.After(after) {
				t.Fatal("verification event must keep its instant; updated_at must use database update time")
			}
		})
	}
}

func TestUpdatedAtUsesWriteTimeAndRollsBackWithRow(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	var original time.Time
	if err := s.pool.QueryRow(ctx, `SELECT updated_at FROM users WHERE id=1`).Scan(&original); err != nil {
		t.Fatal(err)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	// Deliberately separate transaction start from the write. NOW() would be stale.
	if _, err := tx.Exec(ctx, `SELECT pg_sleep(0.02)`); err != nil {
		t.Fatal(err)
	}
	var before, updated, after time.Time
	if err := tx.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `UPDATE users SET login='changed' WHERE id=1 RETURNING updated_at,clock_timestamp()`).Scan(&updated, &after); err != nil {
		t.Fatal(err)
	}
	if updated.Before(before) || updated.After(after) || updated.Equal(original) {
		t.Fatal("updated_at used transaction start instead of the actual write")
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	var restored time.Time
	var login string
	if err := s.pool.QueryRow(ctx, `SELECT login,updated_at FROM users WHERE id=1`).Scan(&login, &restored); err != nil {
		t.Fatal(err)
	}
	if login != "first" || !restored.Equal(original) {
		t.Fatal("rollback did not restore the row and its metadata together")
	}
}

func TestUpdatedAtAccountSelectionAndLogout(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewCharacterRepository(s.pool)
	actor := repositoryActor(t, ctx, s, 1, "metadata-session")
	var previous, before, after, updated time.Time
	if err := s.pool.QueryRow(ctx, `SELECT updated_at,clock_timestamp() FROM users WHERE id=1`).Scan(&previous, &before); err != nil {
		t.Fatal(err)
	}
	characterID := 1
	if _, err := r.SetActive(ctx, actor, &characterID); err != nil {
		t.Fatal(err)
	}
	var selected *int
	if err := s.pool.QueryRow(ctx, `SELECT character_id,updated_at,clock_timestamp() FROM users WHERE id=1`).Scan(&selected, &updated, &after); err != nil {
		t.Fatal(err)
	}
	if selected == nil || *selected != characterID || updated.Equal(previous) || updated.Before(before) || updated.After(after) {
		t.Fatal("character selection did not update account metadata")
	}
	previous = updated
	if _, err := r.SetActive(ctx, actor, &characterID); err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT updated_at FROM users WHERE id=1`).Scan(&updated); err != nil {
		t.Fatal(err)
	}
	if !updated.Equal(previous) {
		t.Fatal("repeating the same character selection changed updated_at")
	}
	if err := s.DeleteSession(ctx, actor.SessionHash); err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT character_id,updated_at FROM users WHERE id=1`).Scan(&selected, &updated); err != nil {
		t.Fatal(err)
	}
	if selected != nil || !updated.After(previous) {
		t.Fatal("logout did not update cleared selection metadata")
	}
	previous = updated
	if err := s.DeleteSession(ctx, actor.SessionHash); err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT updated_at FROM users WHERE id=1`).Scan(&updated); err != nil {
		t.Fatal(err)
	}
	if !updated.Equal(previous) {
		t.Fatal("repeated logout changed account metadata")
	}
}

func TestTimestampMigrationFailureIsAtomic(t *testing.T) {
	s, ctx := characterTestStorageBefore(t, timestampMigration)
	before := timestampSnapshot(t, ctx, s)
	// An existing function with this name makes migration fail after ALTER TABLE.
	mustExec(t, ctx, s, `CREATE FUNCTION set_row_updated_at() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END; $$`)
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, migrationSQL(t, timestampMigration+".up.sql")); err == nil {
		t.Fatal("conflicting migration succeeded")
	}
	if _, err := conn.Exec(ctx, "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	var withoutZone int
	if err := conn.QueryRow(ctx, `SELECT count(*) FROM information_schema.columns
		WHERE table_schema=current_schema() AND data_type='timestamp without time zone'`).Scan(&withoutZone); err != nil || withoutZone != 16 {
		t.Fatalf("failed migration changed column types: %d %v", withoutZone, err)
	}
	if after := timestampSnapshot(t, ctx, s); after != before {
		t.Fatal("failed migration changed stored values")
	}
}

func TestDatabaseTimestampsHaveTimezone(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	var withZone, withoutZone int
	if err := s.pool.QueryRow(ctx, `SELECT
		count(*) FILTER (WHERE data_type='timestamp with time zone'),
		count(*) FILTER (WHERE data_type='timestamp without time zone')
		FROM information_schema.columns WHERE table_schema=current_schema()`).Scan(&withZone, &withoutZone); err != nil {
		t.Fatal(err)
	}
	if withZone != 21 || withoutZone != 0 {
		t.Fatalf("timestamp columns: with zone=%d without zone=%d; want 21/0", withZone, withoutZone)
	}
}

func TestUpdatedAtTracksChangesAndPreservesNoOps(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	insertTimestampCredentials(t, ctx, s)
	changes := map[string]string{
		"users":               "login=login||'_edited'",
		"sessions":            "expires_at=expires_at+interval '1 hour'",
		"email_verifications": "expired_at=expired_at+interval '1 hour'",
		"factions":            "description=coalesce(description,'')||' edited'",
		"classes":             "name=name||' edited'",
		"abilities":           "damage=damage+1",
		"characters":          "experience=experience+1",
	}
	for _, table := range timestampTables {
		t.Run(table, func(t *testing.T) {
			var created, oldUpdated, before, after, newCreated, newUpdated time.Time
			if err := s.pool.QueryRow(ctx, "SELECT created_at,updated_at,clock_timestamp() FROM "+table+" WHERE id=1").Scan(&created, &oldUpdated, &before); err != nil {
				t.Fatal(err)
			}
			if err := s.pool.QueryRow(ctx, "UPDATE "+table+" SET "+changes[table]+" WHERE id=1 RETURNING created_at,updated_at,clock_timestamp()").Scan(&newCreated, &newUpdated, &after); err != nil {
				t.Fatal(err)
			}
			if !newCreated.Equal(created) || newUpdated.Equal(oldUpdated) || newUpdated.Before(before) || newUpdated.After(after) {
				t.Fatalf("metadata did not track the update: created preserved=%v old=%v new=%v bounds=[%v,%v]", newCreated.Equal(created), oldUpdated, newUpdated, before, after)
			}
			var unchanged time.Time
			if err := s.pool.QueryRow(ctx, "UPDATE "+table+" SET id=id WHERE id=1 RETURNING updated_at").Scan(&unchanged); err != nil {
				t.Fatal(err)
			}
			if !unchanged.Equal(newUpdated) {
				t.Fatal("writing identical values changed updated_at")
			}
		})
	}
}
