package db_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gotest/internal/auth"
	"gotest/internal/notification"
)

// This public fixture key never protects application data.
const publicMailQueueKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

func testCipher() *notification.Cipher {
	cipher, err := notification.NewCipher([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		panic(err)
	}
	return cipher
}

var mailFixtureSchemaMu sync.Mutex

// Historical migration fixtures deliberately stop before mail_jobs. Legacy
// repository tests create its additive table only when they need queue writes.
func (s *testStorage) ensureMailQueue(ctx context.Context) error {
	mailFixtureSchemaMu.Lock()
	defer mailFixtureSchemaMu.Unlock()
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT to_regclass(current_schema()||'.mail_jobs') IS NOT NULL`).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	sql, err := os.ReadFile(filepath.Join("..", "..", "migrations", "20261027000200_mail_jobs.up.sql"))
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, string(sql))
	return err
}

// Legacy SQL tests intentionally distinguish arbitrary stored hashes from raw
// service tokens. Give only the new encrypted payload a valid fixture token.
func (s *testStorage) CreateRegistration(ctx context.Context, input auth.Registration) error {
	if err := s.ensureMailQueue(ctx); err != nil {
		return err
	}
	if input.VerificationToken == "" {
		input.VerificationToken = strings.Repeat("1", 64)
	}
	return s.AuthRepository.CreateRegistration(ctx, input)
}

func (s *testStorage) ReplaceVerification(ctx context.Context, userID int, hash string, expires time.Time) error {
	if err := s.ensureMailQueue(ctx); err != nil {
		return err
	}
	return s.AuthRepository.ReplaceVerification(ctx, auth.Verification{UserID: userID, Token: strings.Repeat("2", 64), TokenHash: hash, ExpiresAt: expires})
}

func (s *testStorage) CreatePasswordReset(ctx context.Context, email, hash string, expires time.Time) (bool, error) {
	if err := s.ensureMailQueue(ctx); err != nil {
		return false, err
	}
	message, err := testCipher().Seal(notification.ResetRequest, strings.ToLower(strings.TrimSpace(email)), "", time.Now().Add(time.Hour))
	if err != nil {
		return false, err
	}
	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return false, err
	}
	lease := hex.EncodeToString(nonce[:])
	_, err = s.pool.Exec(ctx, `INSERT INTO mail_jobs(id,kind,key_id,payload,expires_at,attempts,lease_token,leased_until)
		VALUES($1,'reset_request',$2,$3,$4,1,$5,clock_timestamp()+interval '120 seconds')`, message.ID, message.KeyID, message.Ciphertext, message.ExpiresAt, lease)
	if err != nil {
		return false, err
	}
	return s.AuthRepository.CreatePasswordReset(ctx, auth.PasswordReset{Email: email, Token: strings.Repeat("3", 64), TokenHash: hash, ExpiresAt: expires, RequestID: message.ID, RequestLease: lease})
}
