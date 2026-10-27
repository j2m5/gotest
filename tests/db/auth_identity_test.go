package db_test

import (
	"gotest/internal/auth"
	"reflect"
	"testing"
	"time"
)

func TestRegistrationCredentialsAndNullableIdentityRoundTrip(t *testing.T) {
	s, ctx := authTestStorage(t, true)
	registration := auth.Registration{
		Email: "Named@example.test", Login: "named-login", PasswordHash: "password-digest",
		VerificationTokenHash: "verification-digest",
		VerificationExpiresAt: time.Now().Add(2 * time.Hour).Truncate(time.Microsecond),
	}
	if err := s.CreateRegistration(ctx, registration); err != nil {
		t.Fatal(err)
	}
	credentials, err := s.FindCredentialsByEmail(ctx, " NAMED@example.TEST ")
	if err != nil {
		t.Fatal(err)
	}
	if credentials.User.ID <= 0 || credentials.User.Email != registration.Email || credentials.User.Login != registration.Login || credentials.PasswordHash != registration.PasswordHash || credentials.User.CharacterID != nil || credentials.User.EmailVerifiedAt != nil {
		t.Fatalf("registration credentials did not round trip: %+v", credentials.User)
	}
	var storedEmail, storedLogin, storedPassword, storedToken string
	var storedExpiry time.Time
	var roleID int16
	if err := s.pool.QueryRow(ctx, `SELECT u.email,u.login,u.password,v.token,v.expired_at,u.role_id FROM users u JOIN email_verifications v ON v.user_id=u.id WHERE u.id=$1`, credentials.User.ID).Scan(&storedEmail, &storedLogin, &storedPassword, &storedToken, &storedExpiry, &roleID); err != nil {
		t.Fatal(err)
	}
	if storedEmail != registration.Email || storedLogin != registration.Login || storedPassword != registration.PasswordHash || storedToken != registration.VerificationTokenHash || !storedExpiry.Equal(registration.VerificationExpiresAt) || credentials.User.RoleID != roleID {
		t.Fatal("named registration fields or account role were not persisted independently")
	}
	if err := createStoredSession(t, s, ctx, credentials.User.ID, "identity-session", registration.VerificationExpiresAt); err != nil {
		t.Fatal(err)
	}
	session, err := s.FindSession(ctx, "identity-session")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*session.User, credentials.User) || !session.ExpiresAt.Equal(registration.VerificationExpiresAt) {
		t.Fatalf("nullable identity or expiry changed in session: %+v", session)
	}
}

func TestSessionLookupDoesNotDependOnPasswordColumn(t *testing.T) {
	s, ctx := authTestStorage(t, true)
	var id int
	if err := s.pool.QueryRow(ctx, `INSERT INTO users(email,login,password,email_verified_at) VALUES('identity@example.test','identity','private-password-hash',NOW()) RETURNING id`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(time.Hour).Truncate(time.Microsecond)
	if err := createStoredSession(t, s, ctx, id, "session-storage-key", expires); err != nil {
		t.Fatal(err)
	}
	// The fixture owns a disposable schema. Removing credentials from its account
	// table proves that session reads depend only on identity, not on password data.
	if _, err := s.pool.Exec(ctx, `ALTER TABLE users DROP COLUMN password`); err != nil {
		t.Fatal(err)
	}
	session, err := s.FindSession(ctx, "session-storage-key")
	if err != nil {
		t.Fatalf("session lookup must not read credentials: %v", err)
	}
	if session.User.ID != id || session.User.Email != "identity@example.test" || session.User.Login != "identity" || session.User.CharacterID != nil || session.User.EmailVerifiedAt == nil || !session.ExpiresAt.Equal(expires) {
		t.Fatalf("safe account fields lost: %+v", session)
	}
}
