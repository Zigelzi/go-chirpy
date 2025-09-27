package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	t.Run("JWT is made with given details", func(t *testing.T) {
		userUUID := uuid.New()
		tokenSecret := "testing-is-fun"
		tokenString, err := MakeJWT(userUUID, tokenSecret, 200*time.Millisecond)

		if tokenString == "" {
			t.Errorf("token string is empty unexpectedly")
		}
		if err != nil {
			t.Errorf("unexpected error: got [%v] want [nil]", err)
		}
	})
}

func TestUserAuthentication(t *testing.T) {
	userUUID := uuid.New()
	validToken, _ := MakeJWT(userUUID, "secret", time.Hour)
	expiredToken, _ := MakeJWT(userUUID, "secret", -time.Hour)
	tokenWithDifferentIssuer := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJub3QtY2hpcnB5Iiwic3ViIjoiZTJjZGE1ZGEtM2IxNy00MjZhLWE0MDgtZmZmNWI5ZDJjYTZkIiwiZXhwIjoxNzU4OTc0NDIyLCJpYXQiOjE3NTg5NzA4MjJ9.yGu8VCeY8C5xJJi9D1dlSeXb0PFkYB2f1CKLWS9zC48"

	tests := []struct {
		scenario         string
		tokenString      string
		tokenSecret      string
		expectedUserUUID uuid.UUID
		shouldErr        bool
		expectedError    error
	}{
		{
			scenario:         "succeeds with correct token and valid secret",
			tokenString:      validToken,
			tokenSecret:      "secret",
			expectedUserUUID: userUUID,
			shouldErr:        false,
			expectedError:    nil,
		},
		{
			scenario:         "fails when token secret is incorrect",
			tokenString:      validToken,
			tokenSecret:      "weird-secret",
			expectedUserUUID: uuid.Nil,
			shouldErr:        true,
			expectedError:    jwt.ErrSignatureInvalid,
		},
		{
			scenario:         "fails when token is incorrectly formatted",
			tokenString:      "derp.derp.derp",
			tokenSecret:      "secret",
			expectedUserUUID: uuid.Nil,
			shouldErr:        true,
			expectedError:    jwt.ErrTokenMalformed,
		},
		{
			scenario:         "fails when token is empty",
			tokenString:      "",
			tokenSecret:      "secret",
			expectedUserUUID: uuid.Nil,
			shouldErr:        true,
			expectedError:    jwt.ErrTokenMalformed,
		},
		{
			scenario:         "fails when token is expired",
			tokenString:      expiredToken,
			tokenSecret:      "secret",
			expectedUserUUID: uuid.Nil,
			shouldErr:        true,
			expectedError:    jwt.ErrTokenExpired,
		},
		{
			scenario:         "fails when token is not issued by chirpy",
			tokenString:      tokenWithDifferentIssuer,
			tokenSecret:      "secret",
			expectedUserUUID: uuid.Nil,
			shouldErr:        true,
			expectedError:    ErrIncorrectIssuer,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.scenario, func(t *testing.T) {
			actualUserUUID, err := ValidateJWT(testCase.tokenString, testCase.tokenSecret)
			if testCase.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got [nil]")
					return
				}
				if errors.Is(err, testCase.expectedError) == false {
					t.Errorf("Errors don't match: got [%s] want [%s]", err, testCase.expectedError)
					return
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: got [%v] want [nil]", err)
					return
				}

			}

			if actualUserUUID != testCase.expectedUserUUID {
				t.Errorf("User UUIDs don't match: got [%v] want [%s]", actualUserUUID, testCase.expectedUserUUID)
				return
			}
		})

	}
}
