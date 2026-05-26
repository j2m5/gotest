package db

import (
	"context"
	"gotest/internal/models"
)

func (s *Storage) CreateUser(email string, login string, password string) (*models.User, error) {
	query := "INSERT INTO users (email, login, password) VALUES ($1, $2, $3) RETURNING id, role_id, name, email, login, password, email_verified_at, created_at, updated_at"

	row := s.pool.QueryRow(
		context.Background(),
		query,
		email,
		login,
		password,
	)

	user := &models.User{}

	err := row.Scan(
		&user.ID,
		&user.RoleID,
		&user.Name,
		&user.Email,
		&user.Login,
		&user.Password,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Storage) UpdateEmailVerifiedAt(id int) error {
	query := "UPDATE users SET email_verified_at = NOW() WHERE id = $1"

	_, err := s.pool.Exec(context.Background(), query, id)

	return err
}

func (s *Storage) FindUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, role_id, name, email, login, password, email_verified_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	row := s.pool.QueryRow(context.Background(), query, email)

	user := &models.User{}

	err := row.Scan(&user.ID, &user.RoleID, &user.Name, &user.Email, &user.Login, &user.Password, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}
