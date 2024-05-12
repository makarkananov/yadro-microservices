package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"yadro-microservices/internal/core/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO users (username, password, role) VALUES ($1, $2, $3)",
		user.Username,
		user.Password,
		user.Role,
	)

	if err != nil {
		return fmt.Errorf("error saving user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, "SELECT username, password, role FROM users WHERE username = $1", username)
	var user domain.User
	err := row.Scan(&user.Username, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error getting user by username: %w", err)
	}

	return &user, nil
}
