package storage

import (
	"context"
	"database/sql"

	"github.com/AugustSerenity/marketplace/internal/model"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (login, password_hash, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	return s.db.QueryRowContext(
		ctx,
		query,
		user.Login,
		user.PasswordHash,
		user.CreatedAt,
	).Scan(&user.ID)
}

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	var user model.User
	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`
	err := s.db.QueryRowContext(ctx, query, login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
