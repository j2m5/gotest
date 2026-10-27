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
	logger.LogAttrs(context.Background(), slog.LevelError, "cleanup_failed", append(attrs, details...)...)
	return err
}

func run(logger *slog.Logger) error {
	flags := flag.NewFlagSet("cleanup", flag.ContinueOnError)
	// Flag errors can echo arbitrary argument values; log controlled diagnostics.
	flags.SetOutput(io.Discard)
	batch := flags.Int("batch-size", maintenance.DefaultBatchSize, "maximum rows per DELETE (1..10000)")
	timeout := flags.Duration("timeout", time.Minute, "time limit including database connection")
	dryRun := flags.Bool("dry-run", false, "count expired rows without deleting")
	if err := flags.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			flags.SetOutput(os.Stdout)
			fmt.Fprintln(os.Stdout, "Usage: cleanup [-dry-run] [-batch-size 1000] [-timeout 1m]")
			flags.PrintDefaults()
			return nil
		}
		return failure(logger, "arguments", err, slog.String("reason", "invalid arguments; use -help"))
	}
	if flags.NArg() != 0 {
		return failure(logger, "arguments", errors.New("unexpected arguments"), slog.String("reason", "unexpected arguments; use -help"))
	}
	options := maintenance.Options{BatchSize: *batch, DryRun: *dryRun}
	if err := options.Validate(); err != nil {
		return failure(logger, "arguments", err, slog.String("reason", err.Error()))
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
	started := time.Now()
	result, err := maintenance.Run(ctx, pool, options)
	attrs := []slog.Attr{slog.Time("cutoff", result.Cutoff), slog.Int("batch_size", options.BatchSize),
		slog.Int("batches", result.Batches), slog.Bool("dry_run", options.DryRun),
		slog.Float64("duration_ms", float64(time.Since(started))/float64(time.Millisecond))}
	event := "cleanup_completed"
	if options.DryRun {
		event = "cleanup_preview"
		attrs = append(attrs, slog.Int64("sessions_matched", result.Sessions), slog.Int64("verifications_matched", result.Verifications), slog.Int64("password_resets_matched", result.PasswordResets))
	} else {
		attrs = append(attrs, slog.Int64("sessions_deleted", result.Sessions), slog.Int64("verifications_deleted", result.Verifications), slog.Int64("password_resets_deleted", result.PasswordResets))
	}
	if err != nil {
		return failure(logger, "cleanup", err, attrs...)
	}
	logger.LogAttrs(ctx, slog.LevelInfo, event, attrs...)
	return nil
}
