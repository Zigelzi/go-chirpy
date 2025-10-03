package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNoAuthorizationHeader = errors.New("header is missing authorization key")
	ErrNoTokenString         = errors.New("authorization header is missing the token string")
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to has password: %w", err)
	}
	return string(hashedPassword), nil
}

func CheckHashedPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GetBearerToken(headers http.Header) (string, error) {
	authorizationHeader, ok := headers["Authorization"]

	if !ok {
		return "", ErrNoAuthorizationHeader
	}

	if len(authorizationHeader) < 1 {
		return "", ErrNoTokenString
	}
	stringsInHeader := strings.Split(authorizationHeader[0], " ")
	if len(stringsInHeader) < 2 {
		return "", ErrNoTokenString
	}
	tokenString := strings.TrimSpace(stringsInHeader[1])
	return tokenString, nil
}

func SetTokenLifetime(lifetime time.Duration) time.Duration {
	const maxLifetime = time.Hour

	if lifetime > maxLifetime {
		return maxLifetime
	}
	return lifetime
}
