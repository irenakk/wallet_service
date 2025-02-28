package middleware

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
)

var secretKey = []byte("jwt_token_example")

// --- Валидация токена ---
func ValidateToken(tokenString string) (string, error) {
	claims := &jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	return claims.Subject, nil
}

// --- Middleware авторизации ---
type AuthMiddleware struct {
	Next http.Handler
}

func (am *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		http.Error(w, "Invalid token format", http.StatusUnauthorized)
		return
	}

	// Проверяем токен
	username, err := ValidateToken(tokenParts[1])
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Добавляем username в контекст запроса
	ctx := context.WithValue(r.Context(), "username", username)
	am.Next.ServeHTTP(w, r.WithContext(ctx))
}

// --- Функция middleware ---
func Auth(next http.Handler) http.Handler {
	return &AuthMiddleware{Next: next}
}
