package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// getSecretKey retrieves the JWT secret from environment variables to prevent hardcoding sensitive data.
// It falls back to a default key for local development.
func getSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("corehub_secret_key_2026")
	}
	return []byte(secret)
}

// GenerateToken creates a new JWT token for a given username valid for 24 hours.
func GenerateToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.MapClaims{
		"username": username,
		"exp":      expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecretKey())
}

// ParseToken verifies a JWT token string and returns the extracted username if valid.
func ParseToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return getSecretKey(), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["username"].(string), nil
	}

	return "", errors.New("invalid token")
}
