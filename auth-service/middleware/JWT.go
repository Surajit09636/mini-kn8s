package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// UserContextKey is used to store user info in the request context
type contextKey string
const UserContextKey = contextKey("userContext")

// GenerateJWT generates a JWT token for a given user ID and email
func GenerateJWT(userID uint, email string) (string, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return "", errors.New("Secret key is not set")
	}

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   email,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// VerifyToken validates the JWT and returns the custom claims
func VerifyToken(tokenString string) (*JWTClaims, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("Secret key is not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// JWTMiddleware automatically verifies the token and injects claims into the context
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		claims, err := VerifyToken(tokenString)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Inject the claims into the request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		r = r.WithContext(ctx)

		// Call the next handler
		next(w, r)
	}
}