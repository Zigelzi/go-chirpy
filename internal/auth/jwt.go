package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const issuerName = "chirpy"

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	expiresAt := time.Now().UTC().Add(expiresIn)
	jwtExpiresAt := jwt.NewNumericDate(expiresAt)

	jwtIssuedAt := jwt.NewNumericDate(time.Now().UTC())
	claims := &jwt.RegisteredClaims{
		Issuer:    issuerName,
		IssuedAt:  jwtIssuedAt,
		ExpiresAt: jwtExpiresAt,
		Subject:   userID.String(),
	}
	// log.Printf("JWT issued at: %v", jwtIssuedAt)
	// log.Printf("JWT expires in: %v (%v)", expiresAt, expiresIn)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	userId := uuid.UUID{}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("unable to parse the claims: %w", err)
	} else if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		userId, err = uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("unable parse the UUID from the token: %w", err)
		}

		if claims.Issuer != issuerName {
			return uuid.UUID{}, fmt.Errorf("token is not issued by %s: %w", issuerName, err)
		}

	}

	return userId, nil
}
