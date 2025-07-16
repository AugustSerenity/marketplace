// internal/service/auth.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo         repository.UserRepository
	jwtSecretKey string // Секретный ключ для подписи JWT (должен быть в конфиге)
}

func NewAuthService(repo repository.UserRepository, jwtSecretKey string) *AuthService {
	return &AuthService{
		repo:         repo,
		jwtSecretKey: jwtSecretKey,
	}
}

// LoginUser проверяет логин/пароль и возвращает JWT-токен
func (s *AuthService) LoginUser(ctx context.Context, login, password string) (string, error) {
	// 1. Находим пользователя в БД
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", errors.New("user not found")
	}

	// 2. Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	// 3. Генерируем JWT-токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,                               // Идентификатор пользователя
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Срок действия (24 часа)
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecretKey))
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return tokenString, nil
}
