package db

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/notification"
)

// AuthRepository implements auth persistence using the pool shared by the app.
type AuthRepository struct {
	pool   *pgxpool.Pool
	cipher *notification.Cipher
}

func NewAuthRepository(pool *pgxpool.Pool, cipher *notification.Cipher) *AuthRepository {
	return &AuthRepository{pool: pool, cipher: cipher}
}
