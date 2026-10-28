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
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"gotest/internal/backup"
	"gotest/internal/observability"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if run(logger) != nil {
		os.Exit(1)
	}
}
func failure(logger *slog.Logger, stage, reason string, err error, extra ...slog.Attr) error {
	if err == nil {
		err = errors.New(reason)
	}
	attrs := append(observability.ErrorAttrs(err), slog.String("stage", stage), slog.String("reason", reason))
	logger.LogAttrs(context.Background(), slog.LevelError, "dbbackup_failed", append(attrs, extra...)...)
	return err
}
func run(logger *slog.Logger) error {
	args := os.Args[1:]
	help := func() {
		fmt.Fprintln(os.Stdout, "Usage: dbbackup backup|restore -dir PATH [-timeout 5m]\nbackup reads DATABASE_URL; restore requires RESTORE_DATABASE_URL.\nBoth require MAIL_QUEUE_KEY. PostgreSQL tools: PATH or POSTGRES_BIN_DIR.\nRestore only trusted local backups into a separate empty offline database.")
	}
	if len(args) == 1 && (args[0] == "-help" || args[0] == "--help" || args[0] == "-h") {
		help()
		return nil
	}
	if len(args) == 0 || (args[0] != "backup" && args[0] != "restore") {
		return failure(logger, "arguments", "specify backup or restore; use -help", nil)
	}
	action := args[0]
	flags := flag.NewFlagSet("dbbackup "+action, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	directory := flags.String("dir", "", "new backup directory or existing complete backup")
	timeout := flags.Duration("timeout", 5*time.Minute, "operation time limit")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			help()
			flags.SetOutput(os.Stdout)
			flags.PrintDefaults()
			return nil
		}
		return failure(logger, "arguments", "invalid arguments; use -help", err)
	}
	if flags.NArg() != 0 || *directory == "" || *timeout <= 0 {
		return failure(logger, "arguments", "require -dir PATH and a positive timeout; use -help", nil)
	}
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return failure(logger, "environment", "cannot load optional .env file", err)
	}
	setting := "DATABASE_URL"
	if action == "restore" {
		setting = "RESTORE_DATABASE_URL"
	}
	raw := os.Getenv(setting)
	if raw == "" {
		return failure(logger, "configuration", "explicit database URL is required", nil, slog.String("setting", setting))
	}
	if _, err := backup.ParseConnection(raw, os.Environ()); err != nil {
		return failure(logger, "configuration", err.Error(), err)
	}
	key, err := backup.ParseKey(os.Getenv("MAIL_QUEUE_KEY"))
	if err != nil {
		return failure(logger, "configuration", err.Error(), err, slog.String("setting", "MAIL_QUEUE_KEY"))
	}
	interrupt, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(interrupt, *timeout)
	defer cancel()
	options := backup.Options{DatabaseURL: raw, Directory: *directory, Key: key, BinDir: os.Getenv("POSTGRES_BIN_DIR")}
	started := time.Now()
	var manifest backup.Manifest
	if action == "backup" {
		manifest, err = backup.Backup(ctx, options)
	} else {
		manifest, err = backup.Restore(ctx, options)
	}
	if err != nil {
		var problem *backup.Error
		if errors.As(err, &problem) {
			return failure(logger, problem.Stage, problem.Reason, err)
		}
		return failure(logger, action, "operation failed", err)
	}
	logger.LogAttrs(ctx, slog.LevelInfo, action+"_completed", slog.Int64("migration_version", manifest.MigrationVersion), slog.Float64("duration_ms", float64(time.Since(started))/float64(time.Millisecond)))
	return nil
}
