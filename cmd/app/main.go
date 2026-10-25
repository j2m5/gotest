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
	"gotest/internal/routes"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file loaded; using environment variables")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Mail.Driver == "smtp" && (os.Getenv("MAIL_FROM") == "" || cfg.Mail.From == "noreply@example.test") {
		return fmt.Errorf("MAIL_FROM must be set to a real sender for SMTP")
	}
	sender, err := mailer.New(cfg.Mail)
	if err != nil {
		return fmt.Errorf("configure mail: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	startup, cancel := context.WithTimeout(ctx, 10*time.Second)
	pool, err := db.Open(startup, cfg.DatabaseURL)
	cancel()
	if err != nil {
		return err
	}
	defer pool.Close()
	service := auth.NewService(db.NewAuthRepository(pool), sender, auth.Options{SessionTTL: cfg.SessionTTL, VerificationTTL: cfg.VerificationTTL})
	characterService := character.NewService(db.NewCharacterRepository(pool))
	handler, err := routes.New(handlers.NewHandler(service, strings.HasPrefix(cfg.PublicURL, "https://")), handlers.NewCharacterHandler(characterService), middleware.NewMiddleware(service), routes.Options{StaticDir: "frontend/dist", TrustedOrigin: cfg.PublicURL})
	if err != nil {
		return fmt.Errorf("configure routes: %w", err)
	}
	server := &http.Server{Addr: ":" + cfg.AppPort, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second}
	finished := make(chan error, 1)
	go func() { finished <- server.ListenAndServe() }()
	log.Printf("Listening on port %s (mail driver: %s)", cfg.AppPort, cfg.Mail.Driver)
	select {
	case err = <-finished:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err = server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		return nil
	}
}
