package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNoAuthorizationHeader      = errors.New("header is missing authorization key or value")
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
	authorizationHeader := headers.Get("Authorization")

	if authorizationHeader == "" {
		return "", ErrNoAuthorizationHeader
	}
	stringsInHeader := strings.Split(authorizationHeader, " ")
	if stringsInHeader[0] != authorizationType {
		return "", ErrNoAuthorizationType
	}
	if len(stringsInHeader) < 2 {
		return "", ErrNoAuthorizationCredentials
	}
	tokenString := strings.TrimSpace(stringsInHeader[1])
	return tokenString, nil
}

func MakeRefreshToken() (string, error) {
	// Generate 32 byte token that is always different length.
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	hexToken := hex.EncodeToString(tokenBytes)
	return hexToken, nil
}
