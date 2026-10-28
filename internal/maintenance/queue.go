package maintenance

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/notification"
)

// QueueStatusResult describes one read-only snapshot of the mail queue.
// Due is a subset of Pending; the other primary categories are exclusive.
type QueueStatusResult struct {
	Cutoff                                       time.Time
	Total, Pending, Due, Leased, Failed, Expired int64
	OldestPendingAgeSeconds                      *float64
}

// QueueStatus classifies jobs against one database-clock cutoff, without claims,
// retries or cleanup. An active final-attempt lease remains leased until expiry.
func QueueStatus(ctx context.Context, pool *pgxpool.Pool) (QueueStatusResult, error) {
	var result QueueStatusResult
	err := pool.QueryRow(ctx, `WITH cutoff AS MATERIALIZED (
		SELECT clock_timestamp() AS at
	), classified AS MATERIALIZED (
		SELECT created_at, available_at, CASE
			WHEN expires_at <= cutoff.at THEN 'expired'
			WHEN failed_at IS NOT NULL OR (attempts >= $1 AND (leased_until IS NULL OR leased_until <= cutoff.at)) THEN 'failed'
			WHEN leased_until > cutoff.at THEN 'leased'
			ELSE 'pending' END AS state
		FROM mail_jobs CROSS JOIN cutoff
	)
	SELECT cutoff.at, count(classified.state),
		count(*) FILTER (WHERE state='pending'),
		count(*) FILTER (WHERE state='pending' AND available_at <= cutoff.at),
		count(*) FILTER (WHERE state='leased'),
		count(*) FILTER (WHERE state='failed'),
		count(*) FILTER (WHERE state='expired'),
		max(greatest(0, extract(epoch FROM cutoff.at-created_at))) FILTER (WHERE state='pending')
	FROM cutoff LEFT JOIN classified ON true GROUP BY cutoff.at`, notification.MaxAttempts).
		Scan(&result.Cutoff, &result.Total, &result.Pending, &result.Due, &result.Leased, &result.Failed, &result.Expired, &result.OldestPendingAgeSeconds)
	if err != nil {
		return result, fmt.Errorf("read mail queue status: %w", err)
	}
	result.Cutoff = result.Cutoff.UTC()
	return result, nil
}
