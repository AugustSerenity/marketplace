package storage

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/AugustSerenity/marketplace/internal/model"
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
	// Вставляем пользователя и получаем ID
	return s.db.QueryRowContext(
		ctx,
		query,
		user.Login,
		user.PasswordHash,
		user.CreatedAt,
	).Scan(&user.ID)
}
