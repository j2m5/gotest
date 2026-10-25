package db_test

import (
	"context"
	"errors"
	"gotest/internal/db"
	"testing"
)

func TestOpenReturnsConfigurationError(t *testing.T) {
	pool, err := db.Open(context.Background(), "postgres://%zz")
	if err == nil || pool != nil {
		t.Fatalf("invalid connection string: pool=%v err=%v", pool, err)
	}
}

func TestOpenHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	pool, err := db.Open(ctx, "postgres://localhost:1/test")
	if pool != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled open: pool=%v err=%v", pool, err)
	}
}
