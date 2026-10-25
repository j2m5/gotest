package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"gotest/internal/auth"
	"gotest/internal/character"
	"gotest/internal/config"
	"gotest/internal/db"
	"gotest/internal/handlers"
	"gotest/internal/mailer"
	"gotest/internal/middleware"
	"gotest/internal/observability"
	"gotest/internal/routes"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	if err := run(logger); err != nil {
		os.Exit(1)
	}
}

func applicationFailure(logger *slog.Logger, stage string, err error, details ...slog.Attr) error {
	attrs := append(observability.ErrorAttrs(err), slog.String("stage", stage))
	logger.LogAttrs(context.Background(), slog.LevelError, "application_failed", append(attrs, details...)...)
	return err
}

func run(logger *slog.Logger) error {
	if err := godotenv.Load(); err != nil {
		logger.Info("environment_file_not_loaded")
	}
	cfg, err := config.Load()
	if err != nil {
		// Config validation messages contain setting names and constraints, never values.
		return applicationFailure(logger, "configuration", err, slog.String("reason", err.Error()))
	}
	if cfg.Mail.Driver == "smtp" && (os.Getenv("MAIL_FROM") == "" || cfg.Mail.From == "noreply@example.test") {
		return applicationFailure(logger, "mail_configuration", fmt.Errorf("MAIL_FROM must be set to a real sender for SMTP"), slog.String("setting", "MAIL_FROM"))
	}
	sender, err := mailer.New(cfg.Mail)
	if err != nil {
		return applicationFailure(logger, "mail_configuration", err, slog.String("reason", err.Error()))
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	startup, cancel := context.WithTimeout(ctx, 10*time.Second)
	pool, err := db.Open(startup, cfg.DatabaseURL)
	cancel()
	if err != nil {
		return applicationFailure(logger, "database", err)
	}
	defer pool.Close()
	service := auth.NewService(db.NewAuthRepository(pool), sender, auth.Options{SessionTTL: cfg.SessionTTL, VerificationTTL: cfg.VerificationTTL})
	characterService := character.NewService(db.NewCharacterRepository(pool))
	handler, err := routes.New(handlers.NewHandler(service, strings.HasPrefix(cfg.PublicURL, "https://")), handlers.NewCharacterHandler(characterService), middleware.NewMiddleware(service), routes.Options{
		StaticDir: "frontend/dist", TrustedOrigin: cfg.PublicURL, TrustedProxies: cfg.TrustedProxies, Logger: logger,
		Ready: pool.Ping, Lifecycle: ctx,
	})
	if err != nil {
		return applicationFailure(logger, "routes", err)
	}
	server := &http.Server{Addr: ":" + cfg.AppPort, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second,
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError)}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return applicationFailure(logger, "http_listen", err)
	}
	defer listener.Close()
	finished := make(chan error, 1)
	go func() { finished <- server.Serve(listener) }()
	logger.Info("http_started", "port", cfg.AppPort, "mail_driver", cfg.Mail.Driver)
	select {
	case err = <-finished:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return applicationFailure(logger, "http_serve", err)
	case <-ctx.Done():
		logger.Info("http_stopping")
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err = server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			return applicationFailure(logger, "http_shutdown", err)
		}
		logger.Info("http_stopped")
		return nil
	}
}
