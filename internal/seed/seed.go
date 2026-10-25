// Package seed applies a directory of reference data as one transaction.
package seed

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run executes .sql files in filename order. Scripts must not manage transactions
// themselves. The returned filenames are reported only after a successful commit.
func Run(ctx context.Context, pool *pgxpool.Pool, directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read seed directory: %w", err)
	}
	type script struct{ name, sql string }
	var scripts []script
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read seed %s: %w", path, err)
		}
		if strings.TrimSpace(string(contents)) == "" {
			return nil, fmt.Errorf("seed %s is empty", path)
		}
		scripts = append(scripts, script{name: path, sql: string(contents)})
	}
	if len(scripts) == 0 {
		return nil, fmt.Errorf("no .sql seed files in %s", directory)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin seed transaction: %w", err)
	}
	defer func() {
		// Cancellation of the seed must not prevent releasing its transaction.
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	files := make([]string, 0, len(scripts))
	for _, script := range scripts {
		if _, err := tx.Exec(ctx, script.sql); err != nil {
			return nil, fmt.Errorf("seed %s: %w", script.name, err)
		}
		files = append(files, script.name)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit seed transaction: %w", err)
	}
	return files, nil
}
