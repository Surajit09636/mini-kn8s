package utils

import (
	"errors"
	"os"
	"time"
	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	UserID uint `json:"user_id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token for a given user ID and email
func GenerateJWT(userID uint, email string) (string, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return "", errors.New("Secret key is not set")
	}

	claims := JWTClaims{
		UserID: userID,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: email,
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}