package util

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	jwt.RegisteredClaims
}

func GenerateToken(ID string, FullName string) (string, error) {
	claims := Claims{
		ID:       ID,
		FullName: FullName,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "rental-manufacture-api",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func ParseToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return getJWTSecret(), nil
	})
	if err != nil {
		return "", fmt.Errorf("parse token error: %w", err)
	}
	if !token.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", fmt.Errorf("invalid claims type")
	}
	if claims.ID == "" {
		return "", fmt.Errorf("claim 'id' not found")
	}

	return claims.ID, nil
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("default_secret_key_jangan_lupa_ganti")
	}
	return []byte(secret)
}
