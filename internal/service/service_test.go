package service_test

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/AugustSerenity/marketplace/internal/handler/model/ad"
	"github.com/AugustSerenity/marketplace/internal/handler/model/auth"
	"github.com/AugustSerenity/marketplace/internal/model"
	"github.com/AugustSerenity/marketplace/internal/service"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type mockStorage struct {
	CreateUserFunc     func(ctx context.Context, user *model.User) error
	GetUserByLoginFunc func(ctx context.Context, login string) (*model.User, error)
	CreateAdFunc       func(ctx context.Context, ad *model.Ad) error
	GetAdsFunc         func(ctx context.Context, req *ad.ListRequest, userID int64, offset, limit int) ([]*model.AdWithAuthor, error)
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

func (m *mockStorage) GetAds(ctx context.Context, req *ad.ListRequest, userID int64, offset, limit int) ([]*model.AdWithAuthor, error) {
	return m.GetAdsFunc(ctx, req, userID, offset, limit)
}

func TestService_RegisterUser(t *testing.T) {
	tests := []struct {
		name        string
		req         *auth.RegistrationRequest
		mockSetup   func(*mockStorage)
		expectedErr string
	}{
		{
			name: "successful registration",
			req: &auth.RegistrationRequest{
				Login:    "validuser",
				Password: "validpass123",
			},
			mockSetup: func(m *mockStorage) {
				m.CreateUserFunc = func(ctx context.Context, user *model.User) error {
					user.ID = 1
					return nil
				}
			},
		},
		{
			name: "short login",
			req: &auth.RegistrationRequest{
				Login:    "abc",
				Password: "password",
			},
			expectedErr: "login must be at least 4 characters",
		},
		{
			name: "short password",
			req: &auth.RegistrationRequest{
				Login:    "validuser",
				Password: "short",
			},
			expectedErr: "password must be at least 6 characters",
		},
		{
			name: "storage error",
			req: &auth.RegistrationRequest{
				Login:    "validuser",
				Password: "validpass123",
			},
			mockSetup: func(m *mockStorage) {
				m.CreateUserFunc = func(ctx context.Context, user *model.User) error {
					return errors.New("storage error")
				}
			},
			expectedErr: "storage error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStorage{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			s := service.New(mock, "secret")
			resp, err := s.RegisterUser(context.Background(), tt.req)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), resp.ID)
				assert.Equal(t, tt.req.Login, resp.Login)
			}
		})
	}
}

func TestService_LoginUser(t *testing.T) {
	validPassword := "correctpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		login       string
		password    string
		mockSetup   func(*mockStorage)
		expectedErr string
	}{
		{
			name:     "successful login",
			login:    "validuser",
			password: validPassword,
			mockSetup: func(m *mockStorage) {
				m.GetUserByLoginFunc = func(ctx context.Context, login string) (*model.User, error) {
					return &model.User{
						ID:           1,
						Login:        login,
						PasswordHash: string(hashedPassword),
					}, nil
				}
			},
		},
		{
			name:     "user not found",
			login:    "nonexistent",
			password: validPassword,
			mockSetup: func(m *mockStorage) {
				m.GetUserByLoginFunc = func(ctx context.Context, login string) (*model.User, error) {
					return nil, errors.New("user not found")
				}
			},
			expectedErr: "user not found",
		},
		{
			name:     "wrong password",
			login:    "validuser",
			password: "wrongpassword",
			mockSetup: func(m *mockStorage) {
				m.GetUserByLoginFunc = func(ctx context.Context, login string) (*model.User, error) {
					return &model.User{
						ID:           1,
						Login:        login,
						PasswordHash: string(hashedPassword),
					}, nil
				}
			},
			expectedErr: "invalid password",
		},
		{
			name:     "context canceled",
			login:    "validuser",
			password: validPassword,
			mockSetup: func(m *mockStorage) {
				m.GetUserByLoginFunc = func(ctx context.Context, login string) (*model.User, error) {
					return nil, context.Canceled
				}
			},
			expectedErr: "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStorage{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			s := service.New(mock, "secret")
			token, err := s.LoginUser(context.Background(), tt.login, tt.password)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestService_CreateAd(t *testing.T) {
	validReq := ad.CreateRequest{
		Title:       "Valid Title",
		Description: "Valid description",
		ImageURL:    "http://example.com/image.jpg",
		Price:       100,
	}

	tests := []struct {
		name        string
		req         ad.CreateRequest
		userID      int64
		mockSetup   func(*mockStorage)
		expectedErr string
	}{
		{
			name:   "successful creation",
			req:    validReq,
			userID: 1,
			mockSetup: func(m *mockStorage) {
				m.CreateAdFunc = func(ctx context.Context, ad *model.Ad) error {
					ad.ID = 1
					return nil
				}
			},
		},
		{
			name: "invalid title characters",
			req: ad.CreateRequest{
				Title:       "Invalid @#$ Title",
				Description: validReq.Description,
				ImageURL:    validReq.ImageURL,
				Price:       validReq.Price,
			},
			userID:      1,
			expectedErr: "invalid title characters",
		},
		{
			name: "invalid price",
			req: ad.CreateRequest{
				Title:       validReq.Title,
				Description: validReq.Description,
				ImageURL:    validReq.ImageURL,
				Price:       0,
			},
			userID:      1,
			expectedErr: "price must be positive",
		},
		{
			name:   "storage error",
			req:    validReq,
			userID: 1,
			mockSetup: func(m *mockStorage) {
				m.CreateAdFunc = func(ctx context.Context, ad *model.Ad) error {
					return errors.New("storage error")
				}
			},
			expectedErr: "storage error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStorage{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			s := service.New(mock, "secret")
			result, err := s.CreateAd(context.Background(), tt.req, tt.userID)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), result.ID)
				assert.Equal(t, tt.req.Title, result.Title)
				assert.Equal(t, tt.userID, result.AuthorID)
			}
		})
	}
}

func TestService_GetAds(t *testing.T) {
	now := time.Now()
	sampleAds := []*model.AdWithAuthor{
		{
			ID:          1,
			Title:       "Ad 1",
			Description: "Description 1",
			ImageURL:    "http://example.com/1.jpg",
			Price:       100,
			AuthorID:    1,
			CreatedAt:   now,
			AuthorLogin: "user1",
		},
		{
			ID:          2,
			Title:       "Ad 2",
			Description: "Description 2",
			ImageURL:    "http://example.com/2.jpg",
			Price:       200,
			AuthorID:    2,
			CreatedAt:   now.Add(-time.Hour),
			AuthorLogin: "user2",
		},
	}

	tests := []struct {
		name          string
		req           *ad.ListRequest
		userID        int64
		mockSetup     func(*mockStorage)
		expectedAds   []*model.AdWithAuthor
		expectedError string
	}{
		{
			name: "success with pagination",
			req: &ad.ListRequest{
				Page:     2,
				PageSize: 1,
			},
			userID: 1,
			mockSetup: func(m *mockStorage) {
				m.GetAdsFunc = func(ctx context.Context, req *ad.ListRequest, userID int64, offset, limit int) ([]*model.AdWithAuthor, error) {
					assert.Equal(t, 1, offset)
					assert.Equal(t, 1, limit)
					return []*model.AdWithAuthor{sampleAds[1]}, nil
				}
			},
			expectedAds: []*model.AdWithAuthor{sampleAds[1]},
		},
		{
			name: "filter by price",
			req: &ad.ListRequest{
				Page:     1,
				PageSize: 10,
				MinPrice: 150,
				MaxPrice: 250,
			},
			userID: 1,
			mockSetup: func(m *mockStorage) {
				m.GetAdsFunc = func(ctx context.Context, req *ad.ListRequest, userID int64, offset, limit int) ([]*model.AdWithAuthor, error) {
					assert.Equal(t, float64(150), req.MinPrice)
					assert.Equal(t, float64(250), req.MaxPrice)
					return []*model.AdWithAuthor{sampleAds[1]}, nil
				}
			},
			expectedAds: []*model.AdWithAuthor{sampleAds[1]},
		},
		{
			name: "storage error",
			req: &ad.ListRequest{
				Page:     1,
				PageSize: 10,
			},
			userID: 1,
			mockSetup: func(m *mockStorage) {
				m.GetAdsFunc = func(ctx context.Context, req *ad.ListRequest, userID int64, offset, limit int) ([]*model.AdWithAuthor, error) {
					return nil, errors.New("storage error")
				}
			},
			expectedError: "storage error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStorage{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			s := service.New(mock, "secret")
			result, err := s.GetAds(context.Background(), tt.req, tt.userID)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAds, result)
			}
		})
	}
}

func TestService_ParseListRequest(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expected      ad.ListRequest
		expectedError string
	}{
		{
			name:     "default values",
			query:    "",
			expected: ad.ListRequest{Page: 1, PageSize: 10},
		},
		{
			name:     "valid page and page_size",
			query:    "page=2&page_size=20",
			expected: ad.ListRequest{Page: 2, PageSize: 20},
		},
		{
			name:     "sorting parameters",
			query:    "sort_by=price&sort_order=desc",
			expected: ad.ListRequest{Page: 1, PageSize: 10, SortBy: "price", SortOrder: "desc"},
		},
		{
			name:     "price filters",
			query:    "min_price=100&max_price=500",
			expected: ad.ListRequest{Page: 1, PageSize: 10, MinPrice: 100, MaxPrice: 500},
		},
		{
			name:          "invalid page",
			query:         "page=0",
			expectedError: "invalid page value",
		},
		{
			name:          "invalid page_size",
			query:         "page_size=0",
			expectedError: "invalid page_size value",
		},
		{
			name:          "invalid min_price",
			query:         "min_price=abc",
			expectedError: "invalid min_price value",
		},
		{
			name:          "invalid max_price",
			query:         "max_price=abc",
			expectedError: "invalid max_price value",
		},
		{
			name:          "min_price > max_price",
			query:         "min_price=500&max_price=100",
			expectedError: "min_price cannot be greater than max_price",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.New(nil, "secret")
			values, _ := url.ParseQuery(tt.query)
			result, err := s.ParseListRequest(values)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
