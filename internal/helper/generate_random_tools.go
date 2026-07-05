package helper

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateRandomHex(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateRandomChar(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure bytes: %w", err)
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes), nil
}

func GenerateJWTToken(claims jwt.Claims, secretKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate jwt token: %w", err)
	}

	return tokenString, nil
}
