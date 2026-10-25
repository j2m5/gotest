package db

import "github.com/jackc/pgx/v5/pgxpool"

// AuthRepository implements auth persistence using the pool shared by the app.
type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}
