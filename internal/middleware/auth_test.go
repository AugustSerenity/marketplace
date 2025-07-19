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

func TestAuthMiddleware(t *testing.T) {
	secretKey := "mysecret"
	validToken := generateValidToken(secretKey)
	invalidToken := "Bearer invalid-token-here"
	noBearerToken := "invalid-token"
	missingToken := ""

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
