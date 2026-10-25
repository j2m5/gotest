package db_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/auth"
	"gotest/internal/db"
)

// The fixture owns its pool for setup/inspection; the production Storage is
// exercised only through its exported methods.
type testStorage struct {
	*db.Storage
	pool *pgxpool.Pool
}

// Tests never use the application's DATABASE_URL. Every test owns a random schema.
func authTestStorage(t *testing.T, migrate bool) (*testStorage, context.Context) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is unset; PostgreSQL integration test requires an explicit test database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	var nonce [12]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	schema := "auth_test_" + hex.EncodeToString(nonce[:])
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if _, err := admin.Exec(cleanup, "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE"); err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	config.ConnConfig.RuntimeParams["timezone"] = "Europe/Moscow"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	s := &testStorage{Storage: db.NewStorage(pool), pool: pool}
	for _, name := range []string{"000001_create_users.up.sql", "20260514075823_create_sessions.up.sql", "20260516144648_create_email_verifications.up.sql"} {
		executeMigration(t, ctx, s, name)
	}
	if migrate {
		executeMigration(t, ctx, s, "20261025000100_auth_hardening.up.sql")
	}
	return s, ctx
}

func migrationSQL(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func executeMigration(t *testing.T, ctx context.Context, s *testStorage, name string) {
	t.Helper()
	if _, err := s.pool.Exec(ctx, migrationSQL(t, name)); err != nil {
		t.Fatal(err)
	}
}

func TestAuthRegistrationRollbackAndConflicts(t *testing.T) {
	s, ctx := authTestStorage(t, true)
	expires := time.Now().Add(time.Hour)
	if err := s.CreateRegistration(ctx, "Mixed@example.com", "first", "hash", "token", expires); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ email, login, token string }{
		{" mixed@EXAMPLE.com ", "second", "other"},
		{"other@example.com", "first", "other"},
		{"rollback@example.com", "rollback", "token"},
	} {
		if err := s.CreateRegistration(ctx, tc.email, tc.login, "hash", tc.token, expires); !errors.Is(err, auth.ErrConflict) {
			t.Fatalf("want conflict: %v", err)
		}
	}
	if _, err := s.FindUserByEmail(ctx, "rollback@example.com"); !errors.Is(err, auth.ErrNotFound) {
		t.Fatalf("registration was not rolled back: %v", err)
	}
	u, err := s.FindUserByEmail(ctx, " MIXED@example.COM ")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSession(ctx, u.ID, "session", expires); err != nil {
		t.Fatal(err)
	}
	session, err := s.FindSession(ctx, "session")
	if err != nil || session.User.ID != u.ID || session.ExpiresAt.Sub(expires).Abs() > time.Microsecond {
		t.Fatalf("session = %+v, %v", session, err)
	}
	for range 2 {
		if err := s.DeleteSession(ctx, "session"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.FindSession(ctx, "session"); !errors.Is(err, auth.ErrNotFound) {
		t.Fatalf("revoked session: %v", err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := s.FindUserByEmail(canceled, u.Email); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}

func TestAuthVerificationExpiryAndReplacement(t *testing.T) {
	s, ctx := authTestStorage(t, true)
	now := time.Now().Truncate(time.Microsecond)
	if err := s.CreateRegistration(ctx, "user@example.com", "user", "hash", "expired", now); err != nil {
		t.Fatal(err)
	}
	if err := s.ConsumeVerification(ctx, "expired", now); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("exact expiry: %v", err)
	}
	u, err := s.FindUserByEmail(ctx, "user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if u.EmailVerifiedAt != nil {
		t.Fatal("expired token verified account")
	}
	if err := s.ReplaceVerification(ctx, u.ID, "new", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.ConsumeVerification(ctx, "expired", now); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("replaced token: %v", err)
	}
	// Historical data can contain several links for one account; consuming any
	// current link must invalidate every link belonging to that account.
	if _, err := s.pool.Exec(ctx, `INSERT INTO email_verifications(user_id,token,expired_at) VALUES($1,'sibling',$2)`, u.ID, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.ConsumeVerification(ctx, "new", now); err != nil {
		t.Fatal(err)
	}
	var remaining int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM email_verifications WHERE user_id=$1`, u.ID).Scan(&remaining); err != nil || remaining != 0 {
		t.Fatalf("verification links remaining=%d error=%v", remaining, err)
	}
	if err := s.ConsumeVerification(ctx, "new", now); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("token reused: %v", err)
	}
	if err := s.ReplaceVerification(ctx, u.ID, "again", now.Add(time.Hour)); !errors.Is(err, auth.ErrAlreadyVerified) {
		t.Fatalf("verified resend: %v", err)
	}
	u, err = s.FindUserByEmail(ctx, u.Email)
	if err != nil {
		t.Fatal(err)
	}
	if u.EmailVerifiedAt == nil || !u.EmailVerifiedAt.Equal(now) || !u.UpdatedAt.Equal(now) {
		t.Fatalf("verification timestamps: %+v", u)
	}
}

func TestAuthConcurrentVerification(t *testing.T) {
	s, ctx := authTestStorage(t, true)
	now := time.Now()
	if err := s.CreateRegistration(ctx, "race@example.com", "race", "hash", "single", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 2)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 2 {
		wg.Add(1)
		go func() { defer wg.Done(); <-start; results <- s.ConsumeVerification(ctx, "single", now) }()
	}
	close(start)
	wg.Wait()
	close(results)
	success, invalid := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, auth.ErrInvalidToken) {
			invalid++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || invalid != 1 {
		t.Fatalf("success=%d invalid=%d", success, invalid)
	}
}

func TestAuthVerificationRollback(t *testing.T) {
	s, ctx := authTestStorage(t, true)
	now := time.Now()
	if err := s.CreateRegistration(ctx, "rollback@example.com", "rollback", "hash", "retained", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	_, err := s.pool.Exec(ctx, `CREATE FUNCTION reject_verification_delete() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected deletion failure'; END $$;
	CREATE TRIGGER reject_verification_delete BEFORE DELETE ON email_verifications FOR EACH ROW EXECUTE FUNCTION reject_verification_delete()`)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ConsumeVerification(ctx, "retained", now); err == nil {
		t.Fatal("expected deletion failure")
	}
	u, err := s.FindUserByEmail(ctx, "rollback@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if u.EmailVerifiedAt != nil {
		t.Fatal("failed consumption marked user verified")
	}
	if _, err := s.pool.Exec(ctx, "DROP TRIGGER reject_verification_delete ON email_verifications"); err != nil {
		t.Fatal(err)
	}
	if err := s.ConsumeVerification(ctx, "retained", now); err != nil {
		t.Fatalf("failed consumption lost token: %v", err)
	}
}

func TestAuthReplacementRollback(t *testing.T) {
	s, ctx := authTestStorage(t, true)
	now := time.Now()
	if err := s.CreateRegistration(ctx, "first@example.com", "first", "hash", "first-token", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRegistration(ctx, "second@example.com", "second", "hash", "second-token", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	u, err := s.FindUserByEmail(ctx, "first@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ReplaceVerification(ctx, u.ID, "second-token", now.Add(time.Hour)); !errors.Is(err, auth.ErrConflict) {
		t.Fatalf("replacement collision: %v", err)
	}
	if err := s.ConsumeVerification(ctx, "first-token", now); err != nil {
		t.Fatalf("failed replacement lost old token: %v", err)
	}
}

func TestAuthConcurrentResendAndVerification(t *testing.T) {
	s, ctx := authTestStorage(t, true)
	now := time.Now()
	for i := range 10 {
		email, login, old, next := fmt.Sprintf("race%d@example.com", i), fmt.Sprintf("race%d", i), fmt.Sprintf("old%d", i), fmt.Sprintf("new%d", i)
		if err := s.CreateRegistration(ctx, email, login, "hash", old, now.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
		u, err := s.FindUserByEmail(ctx, email)
		if err != nil {
			t.Fatal(err)
		}
		verify, resend := make(chan error, 1), make(chan error, 1)
		start := make(chan struct{})
		go func() { <-start; verify <- s.ConsumeVerification(ctx, old, now) }()
		go func() { <-start; resend <- s.ReplaceVerification(ctx, u.ID, next, now.Add(time.Hour)) }()
		close(start)
		v, r := <-verify, <-resend
		if v == nil {
			if !errors.Is(r, auth.ErrAlreadyVerified) {
				t.Fatalf("verify won; resend=%v", r)
			}
		} else {
			if !errors.Is(v, auth.ErrInvalidToken) || r != nil {
				t.Fatalf("resend won; verify=%v resend=%v", v, r)
			}
			if err := s.ConsumeVerification(ctx, next, now); err != nil {
				t.Fatal(err)
			}
		}
		if err := s.ConsumeVerification(ctx, old, now); !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("old token reused: %v", err)
		}
	}
}

func TestAuthMigrationLegacyTokensAndDown(t *testing.T) {
	s, ctx := authTestStorage(t, false)
	_, err := s.pool.Exec(ctx, `INSERT INTO users(email,login,password) VALUES('Legacy@example.com','legacy','password');
	INSERT INTO sessions(user_id,token,created_at) VALUES(1,'raw-session','2026-10-20 12:00:00');
	INSERT INTO email_verifications(user_id,token,expired_at) VALUES(1,'raw-verification','2026-10-26 12:00:00')`)
	if err != nil {
		t.Fatal(err)
	}
	executeMigration(t, ctx, s, "20261025000100_auth_hardening.up.sql")
	hash := sha256.Sum256([]byte("raw-session"))
	session, err := s.FindSession(ctx, hex.EncodeToString(hash[:]))
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 10, 27, 9, 0, 0, 0, time.UTC); !session.ExpiresAt.Equal(want) {
		t.Fatalf("expiry %v want %v", session.ExpiresAt, want)
	}
	verificationHash := sha256.Sum256([]byte("raw-verification"))
	var expiry time.Time
	if err := s.pool.QueryRow(ctx, "SELECT expired_at FROM email_verifications WHERE token=$1", hex.EncodeToString(verificationHash[:])).Scan(&expiry); err != nil {
		t.Fatal(err)
	}
	if !expiry.Equal(time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)) {
		t.Fatalf("verification expiry %v", expiry)
	}
	executeMigration(t, ctx, s, "20261025000100_auth_hardening.down.sql")
	var users, sessions, tokens int
	if err := s.pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM users),(SELECT count(*) FROM sessions),(SELECT count(*) FROM email_verifications)").Scan(&users, &sessions, &tokens); err != nil {
		t.Fatal(err)
	}
	if users != 1 || sessions != 0 || tokens != 0 {
		t.Fatalf("down counts %d %d %d", users, sessions, tokens)
	}
}

func TestAuthMigrationRejectsDuplicatesAtomically(t *testing.T) {
	s, ctx := authTestStorage(t, false)
	if _, err := s.pool.Exec(ctx, "INSERT INTO users(email,login,password) VALUES('User@example.com','one','hash'),(' user@example.com ','two','hash')"); err != nil {
		t.Fatal(err)
	}
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, migrationSQL(t, "20261025000100_auth_hardening.up.sql")); err == nil {
		t.Fatal("duplicate migration succeeded")
	}
	if _, err := conn.Exec(ctx, "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	var limit int
	if err := conn.QueryRow(ctx, "SELECT character_maximum_length FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='users' AND column_name='email'").Scan(&limit); err != nil {
		t.Fatal(err)
	}
	if limit != 32 {
		t.Fatalf("failed migration changed email limit to %d", limit)
	}
}

func TestAuthMigrationDownRejectsLongEmailsWithoutDataLoss(t *testing.T) {
	s, ctx := authTestStorage(t, true)
	email := "this-email-is-longer-than-thirty-two@example.com"
	if err := s.CreateRegistration(ctx, email, "long", "hash", "retained", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, migrationSQL(t, "20261025000100_auth_hardening.down.sql")); err == nil {
		t.Fatal("down silently shrank long email")
	}
	if _, err := conn.Exec(ctx, "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	var got string
	if err := conn.QueryRow(ctx, "SELECT email FROM users WHERE login='long'").Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != email {
		t.Fatalf("email changed to %q", got)
	}
	var count int
	if err := conn.QueryRow(ctx, "SELECT count(*) FROM email_verifications").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("failed down removed tokens")
	}
}
