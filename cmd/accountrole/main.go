package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
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

func run(logger *slog.Logger) error {
	fail := func(stage string, err error) error {
		logger.LogAttrs(context.Background(), slog.LevelError, "account_role_failed", append(observability.ErrorAttrs(err), slog.String("stage", stage))...)
		return err
	}
	flags := flag.NewFlagSet("accountrole", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	userID := flags.Int("user-id", 0, "account ID to update")
	role := flags.String("role", "", "user, moderator or administrator")
	envFile := flags.String("env-file", ".env", "dotenv file; existing process environment takes precedence")
	if err := flags.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			flags.SetOutput(os.Stdout)
			fmt.Fprintln(os.Stdout, "Usage: accountrole -user-id ID -role user|moderator|administrator [-env-file .env]")
			flags.PrintDefaults()
			return nil
		}
		return fail("arguments", errors.New("invalid arguments; use -help"))
	}
	if flags.NArg() != 0 || *userID <= 0 || int64(*userID) > 2147483647 || !maintenance.ValidAccountRole(*role) {
		return fail("arguments", errors.New("explicit valid user-id and role are required"))
	}
	if *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			explicit := false
			flags.Visit(func(f *flag.Flag) {
				if f.Name == "env-file" {
					explicit = true
				}
			})
			if explicit || !errors.Is(err, os.ErrNotExist) {
				return fail("environment", err)
			}
		}
	}
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		return fail("configuration", errors.New("DATABASE_URL is required"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := db.Open(ctx, dsn)
	if err != nil {
		return fail("database", err)
	}
	defer pool.Close()
	result, err := maintenance.SetAccountRole(ctx, pool, *userID, *role)
	if err != nil {
		return fail("change", err)
	}
	logger.Info("account_role_changed", "user_id", *userID, "previous_role", result.Previous, "role", result.Role, "changed", result.Changed)
	return nil
}
