package model

import (
	"time"
)

type User struct {
	ID           int64     `db:"id"`
	Login        string    `db:"login"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type Ad struct {
	ID          int64     `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	ImageURL    string    `db:"image_url"`
	Price       float64   `db:"price"`
	AuthorID    int64     `db:"author_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type AdWithAuthor struct {
	ID          int64
	Title       string
	Description string
	ImageURL    string
	Price       float64
	AuthorID    int64
	CreatedAt   time.Time
	AuthorLogin string
}
