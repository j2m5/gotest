package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"gotest/internal/db"
	"gotest/internal/maintenance"
	"gotest/internal/observability"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if err := run(logger); err != nil {
		os.Exit(1)
	}
}

func failure(logger *slog.Logger, stage string, err error, details ...slog.Attr) error {
	attrs := append(observability.ErrorAttrs(err), slog.String("stage", stage))
	logger.LogAttrs(context.Background(), slog.LevelError, "mail_queue_status_failed", append(attrs, details...)...)
	return err
}

func run(logger *slog.Logger) error {
	flags := flag.NewFlagSet("mailqueue", flag.ContinueOnError)
	// Flag parsing errors may contain private argument values.
	flags.SetOutput(io.Discard)
	timeout := flags.Duration("timeout", 10*time.Second, "time limit including database connection")
	if err := flags.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			flags.SetOutput(os.Stdout)
			fmt.Fprintln(os.Stdout, "Usage: mailqueue [-timeout 10s]")
			flags.PrintDefaults()
			return nil
		}
		return failure(logger, "arguments", err, slog.String("reason", "invalid arguments; use -help"))
	}
	if flags.NArg() != 0 {
		return failure(logger, "arguments", errors.New("unexpected arguments"), slog.String("reason", "unexpected arguments; use -help"))
	}
	if *timeout <= 0 {
		return failure(logger, "arguments", errors.New("invalid timeout"), slog.String("reason", "timeout must be positive"))
	}
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return failure(logger, "environment", err)
	}
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		return failure(logger, "configuration", errors.New("missing database URL"), slog.String("setting", "DATABASE_URL"))
	}
	interrupt, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(interrupt, *timeout)
	defer cancel()
	connect, cancelConnect := context.WithTimeout(ctx, 10*time.Second)
	pool, err := db.Open(connect, dsn)
	cancelConnect()
	if err != nil {
		return failure(logger, "database", err)
	}
	defer pool.Close()
	result, err := maintenance.QueueStatus(ctx, pool)
	if err != nil {
		return failure(logger, "status", err)
	}
	logger.LogAttrs(ctx, slog.LevelInfo, "mail_queue_status",
		slog.Time("cutoff", result.Cutoff), slog.Int64("total", result.Total),
		slog.Int64("pending", result.Pending), slog.Int64("due", result.Due),
		slog.Int64("leased", result.Leased), slog.Int64("failed", result.Failed),
		slog.Int64("expired", result.Expired), slog.Any("oldest_pending_age_seconds", result.OldestPendingAgeSeconds))
	return nil
}
