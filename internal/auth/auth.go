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
	ErrNoAuthorizationHeader      = errors.New("header is missing authorization key")
	ErrNoAuthorizationCredentials = errors.New("authorization header is missing credentials")
	ErrNoAuthorizationType        = errors.New("authorization header is missing type")
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
	const authorizationType = "Bearer"
	authorizationHeader, ok := headers["Authorization"]

	if !ok {
		return "", ErrNoAuthorizationHeader
	}

	if len(authorizationHeader) < 1 {
		return "", ErrNoAuthorizationCredentials
	}
	stringsInHeader := strings.Split(authorizationHeader[0], " ")
	if stringsInHeader[0] != authorizationType {
		return "", ErrNoAuthorizationType
	}
	if len(stringsInHeader) < 2 {
		return "", ErrNoAuthorizationCredentials
	}
	tokenString := strings.TrimSpace(stringsInHeader[1])
	return tokenString, nil
}

func SetTokenLifetime(requestedLifetime time.Duration) time.Duration {
	const maxLifetime = time.Hour
	const defaultMinLifetime = time.Hour
	const minLifetime = 0

	if requestedLifetime <= minLifetime {
		return defaultMinLifetime
	}
	if requestedLifetime > maxLifetime {
		return maxLifetime
	}
	return requestedLifetime
}
