package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/AugustSerenity/marketplace/internal/handler"
	"github.com/AugustSerenity/marketplace/internal/handler/model/ad"
	"github.com/AugustSerenity/marketplace/internal/handler/model/auth"
	"github.com/AugustSerenity/marketplace/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	RegisterUserFunc     func(ctx context.Context, req *auth.RegistrationRequest) (*auth.RegistrationResponse, error)
	LoginUserFunc        func(ctx context.Context, login, password string) (string, error)
	CreateAdFunc         func(ctx context.Context, req ad.CreateRequest, userID int64) (*model.Ad, error)
	GetAdsFunc           func(ctx context.Context, req *ad.ListRequest, userID int64) ([]*model.AdWithAuthor, error)
	ParseListRequestFunc func(q url.Values) (ad.ListRequest, error)
}

func (m *mockService) RegisterUser(ctx context.Context, req *auth.RegistrationRequest) (*auth.RegistrationResponse, error) {
	return m.RegisterUserFunc(ctx, req)
}

func (m *mockService) LoginUser(ctx context.Context, login, password string) (string, error) {
	return m.LoginUserFunc(ctx, login, password)
}

func (m *mockService) CreateAd(ctx context.Context, req ad.CreateRequest, userID int64) (*model.Ad, error) {
	return m.CreateAdFunc(ctx, req, userID)
}

func (m *mockService) GetAds(ctx context.Context, req *ad.ListRequest, userID int64) ([]*model.AdWithAuthor, error) {
	return m.GetAdsFunc(ctx, req, userID)
}

func (m *mockService) ParseListRequest(q url.Values) (ad.ListRequest, error) {
	return m.ParseListRequestFunc(q)
}

func TestHandler_LoginUser(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		contentType string
		mockReturn  string
		mockError   error
		wantStatus  int
	}{
		{
			name:        "successful login",
			requestBody: `{"login":"user","password":"password123"}`,
			contentType: "application/json",
			mockReturn:  "token123",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "invalid credentials",
			requestBody: `{"login":"user","password":"password123"}`,
			contentType: "application/json",
			mockError:   errors.New("invalid password"),
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "invalid JSON",
			requestBody: `invalid json`,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "missing content type",
			requestBody: `{"login":"user","password":"password123"}`,
			contentType: "",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "empty login or password",
			requestBody: `{"login":"","password":""}`,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{
				LoginUserFunc: func(ctx context.Context, login, password string) (string, error) {
					return tt.mockReturn, tt.mockError
				},
			}

			h := handler.New(mockSvc, "secret")

			req := httptest.NewRequest(http.MethodPost, "/auth-login", strings.NewReader(tt.requestBody))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			h.LoginUser(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp auth.LoginResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, tt.mockReturn, resp.Token)
			}
		})
	}
}

func TestHandler_UserRegistration(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		contentType string
		mockError   error
		wantStatus  int
	}{
		{
			name:        "successful registration",
			requestBody: `{"login":"newuser","password":"password123"}`,
			contentType: "application/json",
			wantStatus:  http.StatusCreated,
		},
		{
			name:        "duplicate user",
			requestBody: `{"login":"existing","password":"password123"}`,
			contentType: "application/json",
			mockError:   errors.New("duplicate key"),
			wantStatus:  http.StatusConflict,
		},
		{
			name:        "invalid JSON",
			requestBody: `invalid json`,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "missing content type",
			requestBody: `{"login":"newuser","password":"password123"}`,
			contentType: "",
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{
				RegisterUserFunc: func(ctx context.Context, req *auth.RegistrationRequest) (*auth.RegistrationResponse, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return &auth.RegistrationResponse{
						ID:    1,
						Login: req.Login,
					}, nil
				},
			}

			h := handler.New(mockSvc, "secret")

			req := httptest.NewRequest(http.MethodPost, "/auth-register", strings.NewReader(tt.requestBody))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			h.UserRegistration(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusCreated {
				var resp auth.RegistrationResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, int64(1), resp.ID)
				assert.Equal(t, "newuser", resp.Login)
			}
		})
	}
}

func TestHandler_CreateAd(t *testing.T) {
	tests := []struct {
		name       string
		request    ad.CreateRequest
		userID     interface{}
		mockError  error
		wantStatus int
	}{
		{
			name: "successful ad creation",
			request: ad.CreateRequest{
				Title:       "Test Ad",
				Description: "Description",
				ImageURL:    "http://example.com/image.jpg",
				Price:       100,
			},
			userID:     int64(1),
			wantStatus: http.StatusCreated,
		},
		{
			name: "unauthorized",
			request: ad.CreateRequest{
				Title: "Test Ad",
			},
			userID:     nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid data",
			request:    ad.CreateRequest{},
			userID:     int64(1),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{
				CreateAdFunc: func(ctx context.Context, req ad.CreateRequest, userID int64) (*model.Ad, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return &model.Ad{
						ID:          1,
						Title:       req.Title,
						Description: req.Description,
						ImageURL:    req.ImageURL,
						Price:       req.Price,
						AuthorID:    userID,
					}, nil
				},
			}

			h := handler.New(mockSvc, "secret")

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/create-ads", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), "userID", tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			h.CreateAd(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestHandler_GetAds(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		userID      interface{}
		mockAds     []*model.AdWithAuthor
		mockError   error
		parseError  error
		wantStatus  int
	}{
		{
			name:        "successful fetch with owner",
			queryParams: "page=1&page_size=10&sort_by=price&sort_order=asc",
			userID:      int64(1),
			mockAds: []*model.AdWithAuthor{
				{
					ID:          1,
					Title:       "Test Ad",
					Description: "Description",
					ImageURL:    "http://example.com/image.jpg",
					Price:       100,
					AuthorID:    1,
					CreatedAt:   time.Now(),
					AuthorLogin: "user1",
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "empty result",
			queryParams: "page=1&page_size=10&sort_by=price&sort_order=asc",
			userID:      int64(1),
			mockAds:     []*model.AdWithAuthor{},
			wantStatus:  http.StatusOK,
		},
		{
			name:        "service error",
			queryParams: "page=1&page_size=10&sort_by=price&sort_order=asc",
			userID:      int64(1),
			mockError:   errors.New("internal"),
			wantStatus:  http.StatusInternalServerError,
		},
		{
			name:        "invalid params",
			queryParams: "page=0&page_size=0",
			userID:      int64(1),
			parseError:  errors.New("invalid params"),
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{
				GetAdsFunc: func(ctx context.Context, req *ad.ListRequest, userID int64) ([]*model.AdWithAuthor, error) {
					return tt.mockAds, tt.mockError
				},
				ParseListRequestFunc: func(q url.Values) (ad.ListRequest, error) {
					if tt.parseError != nil {
						return ad.ListRequest{}, tt.parseError
					}
					return ad.ListRequest{
						Page:      1,
						PageSize:  10,
						SortBy:    "price",
						SortOrder: "asc",
					}, nil
				},
			}

			h := handler.New(mockSvc, "secret")

			req := httptest.NewRequest(http.MethodGet, "/watch-ads?"+tt.queryParams, nil)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), "userID", tt.userID))
			}
			w := httptest.NewRecorder()

			h.GetAds(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestHandler_GetAds_ContentCheck(t *testing.T) {
	mockAds := []*model.AdWithAuthor{
		{
			ID:          1,
			Title:       "Cool Bike",
			Description: "Fast and red",
			ImageURL:    "http://image.jpg",
			Price:       500,
			AuthorID:    42,
			AuthorLogin: "biker",
			CreatedAt:   time.Now(),
		},
	}

	mockSvc := &mockService{
		GetAdsFunc: func(ctx context.Context, req *ad.ListRequest, userID int64) ([]*model.AdWithAuthor, error) {
			return mockAds, nil
		},
		ParseListRequestFunc: func(q url.Values) (ad.ListRequest, error) {
			return ad.ListRequest{Page: 1, PageSize: 10}, nil
		},
	}

	h := handler.New(mockSvc, "secret")

	req := httptest.NewRequest(http.MethodGet, "/watch-ads?page=1&page_size=10", nil)
	req = req.WithContext(context.WithValue(req.Context(), "userID", int64(42))) // same as AuthorID
	w := httptest.NewRecorder()

	h.GetAds(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []ad.ListResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp, 1)

	assert.Equal(t, "Cool Bike", resp[0].Title)
	assert.Equal(t, "biker", resp[0].AuthorLogin)
	assert.True(t, resp[0].IsOwner)
}
