package auth

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	const issuerName = "chirpy"
	jwtIssuedAt := jwt.NewNumericDate(time.Now().UTC())
	expiresAt := time.Now().UTC().Add(expiresIn)
	jwtExpiresAt := jwt.NewNumericDate(expiresAt)
	log.Printf("JWT issued at %v and expires at %v (%v)\n", jwtIssuedAt, jwtExpiresAt, expiresIn)
	claims := &jwt.RegisteredClaims{
		Issuer:    issuerName,
		IssuedAt:  jwtIssuedAt,
		ExpiresAt: jwtExpiresAt,
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateJWT(token, tokenSecret string) (uuid.UUID, error) {
	return uuid.New(), nil
}
