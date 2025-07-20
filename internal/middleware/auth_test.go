package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateValidToken(secretKey string) string {
	claims := jwt.MapClaims{
		"sub": 12345,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		panic(err)
	}
	return tokenString
}

func generateExpiredToken(secretKey string) string {
	claims := jwt.MapClaims{
		"sub": 12345,
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secretKey))
	return tokenString
}

func generateTokenWithStringSub(secretKey string) string {
	claims := jwt.MapClaims{
		"sub": "not-an-integer",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secretKey))
	return tokenString
}

func generateTokenWithWrongSignature(secretKey string) string {
	wrongKey := "wrongsecret"
	claims := jwt.MapClaims{
		"sub": 12345,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(wrongKey))
	return tokenString
}

func TestAuthMiddleware(t *testing.T) {
	secretKey := "mysecret"
	validToken := generateValidToken(secretKey)
	invalidToken := "Bearer invalid-token-here"
	noBearerToken := "invalid-token"
	missingToken := ""

	// Создание тестового обработчика
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userID").(int64)
		if !ok || userID == 0 {
			t.Error("Expected userID to be set in context, but got nil or 0")
		}
		w.WriteHeader(http.StatusOK)
	})

	authHandler := AuthMiddleware(secretKey)(handler)

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{"Valid Token", "Bearer " + validToken, http.StatusOK},
		{"Invalid Token", invalidToken, http.StatusUnauthorized},
		{"No Bearer Token", noBearerToken, http.StatusUnauthorized},
		{"Missing Token", missingToken, http.StatusUnauthorized},
		{"Expired Token", "Bearer " + generateExpiredToken(secretKey), http.StatusUnauthorized},
		{"Token With String Sub", "Bearer " + generateTokenWithStringSub(secretKey), http.StatusUnauthorized},
		{"Token With Wrong Signature", "Bearer " + generateTokenWithWrongSignature(secretKey), http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/protected", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			resp := httptest.NewRecorder()
			authHandler.ServeHTTP(resp, req)

			if resp.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, resp.Code)
			}
		})
	}
}

func TestOptionalAuthMiddleware(t *testing.T) {
	secretKey := "mysecret"
	validToken := generateValidToken(secretKey)
	invalidToken := "Bearer invalid.token.here"

	tests := []struct {
		name         string
		token        string
		expectUserID bool
		expectStatus int
	}{
		{
			name:         "No Token Provided",
			token:        "",
			expectUserID: false,
			expectStatus: http.StatusOK,
		},
		{
			name:         "Valid Token Provided",
			token:        "Bearer " + validToken,
			expectUserID: true,
			expectStatus: http.StatusOK,
		},
		{
			name:         "Invalid Token Provided",
			token:        invalidToken,
			expectUserID: false,
			expectStatus: http.StatusOK, // still OK because it's optional
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, ok := r.Context().Value("userID").(int64)
				if ok != tt.expectUserID {
					t.Errorf("expected userID presence: %v, got: %v", tt.expectUserID, ok)
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/optional", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			w := httptest.NewRecorder()
			OptionalAuthMiddleware(secretKey)(handler).ServeHTTP(w, req)

			if w.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, w.Code)
			}
		})
	}
}
