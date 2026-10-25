package db_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/auth"
	"gotest/internal/character"
	"gotest/internal/db"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/routes"
	"gotest/tests/testenv"
)

func TestReadinessTracksRealDatabaseAvailability(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, testenv.DatabaseURL(t))
	if err != nil {
		t.Fatal("cannot configure dedicated test pool")
	}
	t.Cleanup(pool.Close)
	service := auth.NewService(db.NewAuthRepository(pool), &flowMail{}, auth.Options{})
	h, err := routes.New(handlers.NewHandler(service, false), handlers.NewCharacterHandler(character.NewService(db.NewCharacterRepository(pool))), middleware.NewMiddleware(service), routes.Options{
		StaticDir: t.TempDir(), Ready: pool.Ping, Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	flowRequest(t, h, "GET", "/readyz", "", nil, 200)
	pool.Close()
	flowRequest(t, h, "GET", "/readyz", "", nil, 503)
	flowRequest(t, h, "GET", "/healthz", "", nil, 200)
}
