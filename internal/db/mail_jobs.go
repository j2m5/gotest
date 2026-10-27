package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/notification"
)

// MailQueue persists encrypted inbox and outbox jobs in the application database.
type MailQueue struct{ pool *pgxpool.Pool }

var _ notification.Store = (*MailQueue)(nil)

func NewMailQueue(pool *pgxpool.Pool) *MailQueue { return &MailQueue{pool: pool} }

func validMailHex(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

// enqueueMail is shared with auth transactions so changing an account and
// creating its outgoing notification either both commit or both roll back.
func enqueueMail(ctx context.Context, tx pgx.Tx, message notification.Message, userID *int) error {
	if !validMailHex(message.ID) || !validMailHex(message.KeyID) || len(message.Ciphertext) == 0 || message.ExpiresAt.IsZero() {
		return errors.New("invalid encrypted mail metadata")
	}
	switch message.Kind {
	case notification.ResetRequest:
		if userID != nil {
			return errors.New("incoming reset request cannot have an account owner")
		}
	case notification.Verification, notification.PasswordReset, notification.PasswordChanged:
		if userID == nil || *userID <= 0 {
			return errors.New("outgoing mail requires an account owner")
		}
	default:
		return errors.New("invalid mail kind")
	}
	result, err := tx.Exec(ctx, `INSERT INTO mail_jobs (id,kind,key_id,payload,user_id,expires_at)
		SELECT $1,$2,$3,$4,$5,$6 WHERE $6::timestamptz > clock_timestamp()`,
		message.ID, string(message.Kind), message.KeyID, message.Ciphertext, userID, message.ExpiresAt)
	if err != nil {
		return fmt.Errorf("enqueue encrypted mail: %w", err)
	}
	if result.RowsAffected() != 1 {
		return errors.New("encrypted mail has expired")
	}
	return nil
}

// EnqueueResetRequest accepts intake before any account lookup. The advisory
// lock serializes the capacity check and insert across processes.
func (q *MailQueue) EnqueueResetRequest(ctx context.Context, message notification.Message) error {
	if message.Kind != notification.ResetRequest {
		return errors.New("expected reset request")
	}
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(716183429105)`); err != nil {
		return err
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM mail_jobs WHERE kind='reset_request'
		AND failed_at IS NULL AND expires_at > clock_timestamp()`).Scan(&count); err != nil {
		return err
	}
	if count >= notification.MaxResetRequests {
		return notification.ErrQueueFull
	}
	if err := enqueueMail(ctx, tx, message, nil); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ValidateKey prevents a changed configuration key from turning durable jobs
// into unreadable failures. Retained, unexpired failures need the same key too.
func (q *MailQueue) ValidateKey(ctx context.Context, keyID string) error {
	if !validMailHex(keyID) {
		return errors.New("invalid mail queue key identifier")
	}
	var mismatch bool
	if err := q.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM mail_jobs
		WHERE expires_at > clock_timestamp() AND key_id <> $1)`, keyID).Scan(&mismatch); err != nil {
		return err
	}
	if mismatch {
		return errors.New("mail queue contains jobs encrypted with another key")
	}
	return nil
}

func (q *MailQueue) Claim(ctx context.Context) (*notification.Job, error) {
	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("create mail lease: %w", err)
	}
	job := &notification.Job{}
	// A crash on the final attempt leaves no retry available. Retire that job
	// only after its lease expires, preserving the current owner's right to finish.
	err := q.pool.QueryRow(ctx, `WITH exhausted AS (
		SELECT id FROM mail_jobs WHERE failed_at IS NULL AND attempts >= $1
		AND (leased_until IS NULL OR leased_until <= clock_timestamp())
		FOR UPDATE SKIP LOCKED
	), retired AS (
		UPDATE mail_jobs j SET failed_at=clock_timestamp(),lease_token=NULL,leased_until=NULL
		FROM exhausted e WHERE j.id=e.id
	), candidate AS (
		SELECT id FROM mail_jobs WHERE failed_at IS NULL AND expires_at > clock_timestamp()
		AND available_at <= clock_timestamp() AND attempts < $1
		AND (leased_until IS NULL OR leased_until <= clock_timestamp())
		ORDER BY available_at,created_at,id FOR UPDATE SKIP LOCKED LIMIT 1
	) UPDATE mail_jobs j SET lease_token=$2,
		leased_until=clock_timestamp()+$3 * interval '1 microsecond',attempts=j.attempts+1
		FROM candidate c WHERE j.id=c.id
		RETURNING j.id,j.kind,j.key_id,j.payload,j.expires_at,j.lease_token,j.attempts,j.created_at`,
		notification.MaxAttempts, hex.EncodeToString(nonce[:]), notification.LeaseDuration.Microseconds()).Scan(
		&job.ID, &job.Kind, &job.KeyID, &job.Ciphertext, &job.ExpiresAt, &job.Lease, &job.Attempts, &job.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("claim encrypted mail: %w", err)
	}
	return job, nil
}

func (q *MailQueue) Owns(ctx context.Context, job notification.Job) (bool, error) {
	var owns bool
	err := q.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM mail_jobs
		WHERE id=$1 AND lease_token=$2 AND leased_until > clock_timestamp()
		AND expires_at > clock_timestamp() AND failed_at IS NULL)`, job.ID, job.Lease).Scan(&owns)
	return owns, err
}

func (q *MailQueue) Complete(ctx context.Context, job notification.Job) error {
	result, err := q.pool.Exec(ctx, `DELETE FROM mail_jobs WHERE id=$1 AND lease_token=$2
		AND leased_until > clock_timestamp()`, job.ID, job.Lease)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return notification.ErrLeaseLost
	}
	return nil
}

func (q *MailQueue) Retry(ctx context.Context, job notification.Job, delay time.Duration, permanent bool) error {
	if delay < 0 {
		delay = 0
	}
	result, err := q.pool.Exec(ctx, `UPDATE mail_jobs SET
		failed_at=CASE WHEN $3 OR attempts >= $4 THEN clock_timestamp() ELSE NULL END,
		available_at=LEAST(expires_at,clock_timestamp()+$5 * interval '1 microsecond'),
		lease_token=NULL,leased_until=NULL
		WHERE id=$1 AND lease_token=$2 AND leased_until > clock_timestamp()`,
		job.ID, job.Lease, permanent, notification.MaxAttempts, delay.Microseconds())
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return notification.ErrLeaseLost
	}
	return nil
}

// consumeResetRequest must run before looking up the account. Deletion is
// restored on rollback; after commit, a replay cannot issue another reset token.
func consumeResetRequest(ctx context.Context, tx pgx.Tx, jobID, lease string) error {
	result, err := tx.Exec(ctx, `DELETE FROM mail_jobs WHERE id=$1 AND kind='reset_request'
		AND lease_token=$2 AND leased_until > clock_timestamp()
		AND expires_at > clock_timestamp()`, jobID, lease)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return notification.ErrLeaseLost
	}
	return nil
}
