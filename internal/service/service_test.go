package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AugustSerenity/marketplace/internal/handler/model/ad"
	"github.com/AugustSerenity/marketplace/internal/model"
	"github.com/AugustSerenity/marketplace/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockStorage struct {
	CreateUserFunc     func(ctx context.Context, user *model.User) error
	GetUserByLoginFunc func(ctx context.Context, login string) (*model.User, error)
	CreateAdFunc       func(ctx context.Context, ad *model.Ad) error
	GetAdsFunc         func(ctx context.Context, req *ad.ListRequest, userID int64) ([]*model.AdWithAuthor, error)
}

func (m *mockStorage) CreateUser(ctx context.Context, user *model.User) error {
	return m.CreateUserFunc(ctx, user)
}

func (m *mockStorage) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	return m.GetUserByLoginFunc(ctx, login)
}

func (m *mockStorage) CreateAd(ctx context.Context, ad *model.Ad) error {
	return m.CreateAdFunc(ctx, ad)
}

func (m *mockStorage) GetAds(ctx context.Context, req *ad.ListRequest, userID int64) ([]*model.AdWithAuthor, error) {
	return m.GetAdsFunc(ctx, req, userID)
}

func TestService_RegistrationUser_Success(t *testing.T) {
	called := false
	mockStorage := &mockStorage{
		CreateUserFunc: func(ctx context.Context, user *model.User) error {
			called = true
			user.ID = 1
			return nil
		},
	}

	s := service.New(mockStorage, "secret")

	user := &model.User{
		Login:        "testuser",
		PasswordHash: "hashedpassword",
	}

	err := s.RegistrationUser(context.Background(), user)

	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, int64(1), user.ID)
}

func TestService_LoginUser_Success(t *testing.T) {
	password := "securePassword123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockStorage := &mockStorage{
		GetUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			return &model.User{
				ID:           1,
				Login:        "testuser",
				PasswordHash: string(hashedPassword),
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	s := service.New(mockStorage, "mySecretKey")

	token, err := s.LoginUser(context.Background(), "testuser", password)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims := &jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("mySecretKey"), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
	assert.Equal(t, float64(1), (*claims)["sub"])
}

func TestService_LoginUser_ContextCancellation(t *testing.T) {
	mockStorage := &mockStorage{
		GetUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			select {
			case <-ctx.Done():
				return nil, context.Canceled
			default:
				t.Error("should not reach here")
				return nil, nil
			}
		},
	}

	s := service.New(mockStorage, "secret")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := s.LoginUser(ctx, "testuser", "password")
	assert.ErrorIs(t, err, context.Canceled)
}

func TestService_LoginUser_WrongPassword(t *testing.T) {
	correctPassword := "securePassword123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	mockStorage := &mockStorage{
		GetUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			return &model.User{
				ID:           1,
				Login:        "testuser",
				PasswordHash: string(hashedPassword),
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	s := service.New(mockStorage, "mySecretKey")

	_, err := s.LoginUser(context.Background(), "testuser", "wrongPassword")
	assert.EqualError(t, err, "invalid password")
}

func TestService_LoginUser_UserNotFound(t *testing.T) {
	mockStorage := &mockStorage{
		GetUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			return nil, errors.New("not found")
		},
	}

	s := service.New(mockStorage, "mySecretKey")

	_, err := s.LoginUser(context.Background(), "unknown_user", "anyPassword")
	assert.EqualError(t, err, "user not found")
}

func TestService_CreateAd_Success(t *testing.T) {
	called := false
	mockStorage := &mockStorage{
		CreateAdFunc: func(ctx context.Context, ad *model.Ad) error {
			called = true
			ad.ID = 1
			return nil
		},
	}

	s := service.New(mockStorage, "secret")

	testAd := &model.Ad{
		Title:       "Test Ad",
		Description: "Test Description",
		Price:       100,
	}

	err := s.CreateAd(context.Background(), testAd)

	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, int64(1), testAd.ID)
}

func TestService_CreateAd_ContextCancellation(t *testing.T) {
	mockStorage := &mockStorage{
		CreateAdFunc: func(ctx context.Context, ad *model.Ad) error {
			select {
			case <-ctx.Done():
				return context.Canceled
			default:
				t.Error("should not reach here")
				return nil
			}
		},
	}

	s := service.New(mockStorage, "secret")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.CreateAd(ctx, &model.Ad{
		Title:       "Test",
		Description: "Test",
		Price:       100,
	})
	assert.ErrorIs(t, err, context.Canceled)
}

func TestService_GetAds_Success(t *testing.T) {
	expectedAds := []*model.AdWithAuthor{
		{
			ID:          1,
			Title:       "Test Ad 1",
			Description: "Description 1",
			ImageURL:    "http://example.com/image1.jpg",
			Price:       100,
			AuthorID:    1,
			CreatedAt:   time.Now(),
			AuthorLogin: "user1",
		},
	}

	mockStorage := &mockStorage{
		GetAdsFunc: func(ctx context.Context, req *ad.ListRequest, userID int64) ([]*model.AdWithAuthor, error) {
			return expectedAds, nil
		},
	}

	s := service.New(mockStorage, "secret")

	ads, err := s.GetAds(context.Background(), &ad.ListRequest{
		Page:     1,
		PageSize: 10,
	}, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedAds, ads)
}

func TestService_CreateAd_EmptyTitle(t *testing.T) {
	mockStorage := &mockStorage{
		CreateAdFunc: func(ctx context.Context, ad *model.Ad) error {
			t.Error("storage should not be called when title is empty")
			return nil
		},
	}

	s := service.New(mockStorage, "secret")

	err := s.CreateAd(context.Background(), &model.Ad{
		Title:       "",
		Description: "Valid description",
		Price:       100,
	})

	assert.EqualError(t, err, "title cannot be empty")
}

func TestService_CreateAd_InvalidPrice(t *testing.T) {
	mockStorage := &mockStorage{
		CreateAdFunc: func(ctx context.Context, ad *model.Ad) error {
			t.Error("storage should not be called when price is invalid")
			return nil
		},
	}

	s := service.New(mockStorage, "secret")

	tests := []struct {
		name  string
		price float64
	}{
		{"zero price", 0},
		{"negative price", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.CreateAd(context.Background(), &model.Ad{
				Title:       "Valid title",
				Description: "Valid description",
				Price:       tt.price,
			})

			assert.EqualError(t, err, "price must be positive")
		})
	}
}
