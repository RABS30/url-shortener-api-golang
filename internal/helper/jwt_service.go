package helper

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWTToken(claims jwt.Claims, secretKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate jwt token: %w", err)
	}

	return tokenString, nil
}


