package storage

import (
	"context"
	"database/sql"

	"github.com/AugustSerenity/marketplace/internal/handler/model/ad"
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

func (s *Storage) CreateAd(ctx context.Context, ad *model.Ad) error {
	query := `
		INSERT INTO ads (title, description, image_url, price, author_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	return s.db.QueryRowContext(
		ctx,
		query,
		ad.Title,
		ad.Description,
		ad.ImageURL,
		ad.Price,
		ad.AuthorID,
		ad.CreatedAt,
	).Scan(&ad.ID)
}

func (s *Storage) GetAds(ctx context.Context, req *ad.ListRequest, userID int64) ([]*model.AdWithAuthor, error) {
	query := `
        SELECT 
            a.id, 
            a.title, 
            a.description, 
            a.image_url, 
            a.price, 
            a.author_id,
            a.created_at,
            u.login as author_login
        FROM ads a
        JOIN users u ON a.author_id = u.id
        WHERE ($1 = 0 OR a.price >= $1)
        AND ($2 = 0 OR a.price <= $2)
        ORDER BY 
            CASE WHEN $3 = 'price' AND $4 = 'asc' THEN a.price END ASC,
            CASE WHEN $3 = 'price' AND $4 = 'desc' THEN a.price END DESC,
            CASE WHEN $3 = 'created_at' AND $4 = 'asc' THEN a.created_at END ASC,
            CASE WHEN $3 = 'created_at' AND $4 = 'desc' THEN a.created_at END DESC
        LIMIT $5 OFFSET $6
    `

	offset := (req.Page - 1) * req.PageSize
	rows, err := s.db.QueryContext(
		ctx,
		query,
		req.MinPrice,
		req.MaxPrice,
		req.SortBy,
		req.SortOrder,
		req.PageSize,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ads []*model.AdWithAuthor
	for rows.Next() {
		var ad model.AdWithAuthor
		if err := rows.Scan(
			&ad.ID,
			&ad.Title,
			&ad.Description,
			&ad.ImageURL,
			&ad.Price,
			&ad.AuthorID,
			&ad.CreatedAt,
			&ad.AuthorLogin,
		); err != nil {
			return nil, err
		}
		ads = append(ads, &ad)
	}

	return ads, nil
}
