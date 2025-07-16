package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Добавляем ID пользователя в контекст
			claims := token.Claims.(jwt.MapClaims)
			userID := claims["sub"].(float64) // Приводим к нужному типу
			ctx := context.WithValue(r.Context(), "userID", uint64(userID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
