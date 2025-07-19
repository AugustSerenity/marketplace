package service

import (
	"context"
	"errors"
	"time"

	"github.com/AugustSerenity/marketplace/internal/handler/model/ad"
	"github.com/AugustSerenity/marketplace/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	storage Storage
	secret  string
}

func New(st Storage, secret string) *Service {
	return &Service{
		storage: st,
		secret:  secret,
	}
}

func (s *Service) RegistrationUser(ctx context.Context, user *model.User) error {
	return s.storage.CreateUser(ctx, user)
}

func (s *Service) LoginUser(ctx context.Context, login, password string) (string, error) {

	user, err := s.storage.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return "", err
		}
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return tokenString, nil
}

func (s *Service) CreateAd(ctx context.Context, ad *model.Ad) error {
	if ad.Title == "" {
		return errors.New("title cannot be empty")
	}
	if ad.Price <= 0 {
		return errors.New("price must be positive")
	}
	return s.storage.CreateAd(ctx, ad)
}

func (s *Service) GetAds(ctx context.Context, req *ad.ListRequest, userID int64) ([]*model.AdWithAuthor, error) {
	return s.storage.GetAds(ctx, req, userID)
}
