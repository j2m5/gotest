// Package maintenance contains bounded database maintenance jobs.
package maintenance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DefaultBatchSize = 1000
const MaxBatchSize = 10000

type Options struct {
	BatchSize int
	DryRun    bool
}

// Result counts confirmed deletions, or matching rows when DryRun is enabled.
// On a connection failure an unacknowledged statement's commit can be unknown.
type Result struct {
	Cutoff                  time.Time
	Sessions, Verifications int64
	PasswordResets          int64
	Batches                 int
}

func (o Options) Validate() error {
	if o.BatchSize < 1 || o.BatchSize > MaxBatchSize {
		return errors.New("batch-size must be between 1 and 10000")
	}
	return nil
}

// Run uses one database-clock cutoff for the entire pass. Each DELETE is its own
// transaction: cancellation retains earlier committed batches. Locked rows are
// skipped and may be picked up by a later run; accounts/characters are not changed.
func Run(ctx context.Context, pool *pgxpool.Pool, options Options) (Result, error) {
	var result Result
	if err := options.Validate(); err != nil {
		return result, err
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if err := pool.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&result.Cutoff); err != nil {
		return result, fmt.Errorf("read cleanup cutoff: %w", err)
	}
	result.Cutoff = result.Cutoff.UTC()
	if options.DryRun {
		err := pool.QueryRow(ctx, `SELECT
			(SELECT count(*) FROM sessions WHERE expires_at <= $1),
			(SELECT count(*) FROM email_verifications WHERE expired_at <= $1),
			(SELECT count(*) FROM password_resets WHERE expires_at <= $1)`, result.Cutoff).
			Scan(&result.Sessions, &result.Verifications, &result.PasswordResets)
		if err != nil {
			return result, fmt.Errorf("preview cleanup: %w", err)
		}
		return result, nil
	}
	jobs := []struct {
		query, name string
		count       *int64
		more        bool
	}{
		{`WITH expired AS (
			SELECT id FROM sessions WHERE expires_at <= $1 ORDER BY expires_at,id
			LIMIT $2 FOR UPDATE SKIP LOCKED
		) DELETE FROM sessions USING expired WHERE sessions.id=expired.id`, "sessions", &result.Sessions, true},
		{`WITH expired AS (
			SELECT id FROM email_verifications WHERE expired_at <= $1 ORDER BY expired_at,id
			LIMIT $2 FOR UPDATE SKIP LOCKED
		) DELETE FROM email_verifications USING expired WHERE email_verifications.id=expired.id`, "verifications", &result.Verifications, true},
		{`WITH expired AS (
			SELECT id FROM password_resets WHERE expires_at <= $1 ORDER BY expires_at,id
			LIMIT $2 FOR UPDATE SKIP LOCKED
		) DELETE FROM password_resets USING expired WHERE password_resets.id=expired.id`, "password resets", &result.PasswordResets, true},
	}
	for jobs[0].more || jobs[1].more || jobs[2].more {
		for i := range jobs {
			job := &jobs[i]
			if !job.more {
				continue
			}
			tag, err := pool.Exec(ctx, job.query, result.Cutoff, options.BatchSize)
			if err != nil {
				return result, fmt.Errorf("clean expired %s: %w", job.name, err)
			}
			count := tag.RowsAffected()
			*job.count += count
			if count > 0 {
				result.Batches++
			}
			job.more = count == int64(options.BatchSize)
		}
	}
	return result, nil
}
