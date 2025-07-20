package service

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/AugustSerenity/marketplace/internal/handler/model/ad"
	"github.com/AugustSerenity/marketplace/internal/handler/model/auth"
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

func (s *Service) RegisterUser(ctx context.Context, req *auth.RegistrationRequest) (*auth.RegistrationResponse, error) {
	if len(req.Login) < 4 {
		return nil, errors.New("login must be at least 4 characters")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := model.User{
		Login:        req.Login,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	if err := s.storage.CreateUser(ctx, &user); err != nil {
		return nil, err
	}

	return &auth.RegistrationResponse{
		ID:    user.ID,
		Login: user.Login,
	}, nil
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

func (s *Service) CreateAd(ctx context.Context, req ad.CreateRequest, userID int64) (*model.Ad, error) {
	invalidTitleRegex := `[^a-zA-Z0-9\s]`
	if matched, _ := regexp.MatchString(invalidTitleRegex, req.Title); matched {
		return nil, errors.New("invalid title characters")
	}

	if req.Price <= 0 {
		return nil, errors.New("price must be positive")
	}

	ad := &model.Ad{
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		AuthorID:    userID,
		CreatedAt:   time.Now(),
	}

	if err := s.storage.CreateAd(ctx, ad); err != nil {
		return nil, err
	}

	return ad, nil
}

func (s *Service) GetAds(ctx context.Context, req *ad.ListRequest, userID int64) ([]*model.AdWithAuthor, error) {

	offset := (req.Page - 1) * req.PageSize
	limit := req.PageSize

	ads, err := s.storage.GetAds(
		ctx,
		req, userID, offset, limit,
	)
	if err != nil {
		return nil, err
	}

	return ads, nil
}

func (s *Service) ParseListRequest(q url.Values) (ad.ListRequest, error) {
	req := ad.ListRequest{
		Page:     1,
		PageSize: 10,
	}

	if val := q.Get("page"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			req.Page = parsed
		} else {
			return req, errors.New("invalid page value")
		}
	}

	if val := q.Get("page_size"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			req.PageSize = parsed
		} else {
			return req, errors.New("invalid page_size value")
		}
	}

	req.SortBy = q.Get("sort_by")
	req.SortOrder = q.Get("sort_order")

	if val := q.Get("min_price"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil && parsed >= 0 {
			req.MinPrice = parsed
		} else {
			return req, errors.New("invalid min_price value")
		}
	}

	if val := q.Get("max_price"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil && parsed >= 0 {
			req.MaxPrice = parsed
		} else {
			return req, errors.New("invalid max_price value")
		}
	}

	if req.MinPrice > req.MaxPrice && req.MaxPrice != 0 {
		return req, errors.New("min_price cannot be greater than max_price")
	}

	return req, nil
}
