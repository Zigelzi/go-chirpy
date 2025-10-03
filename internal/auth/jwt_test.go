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

func TestAuthenticationTokenValidation(t *testing.T) {
	userUUID := uuid.New()
	validToken, _ := MakeJWT(userUUID, "secret", time.Hour)
	expiredToken, _ := MakeJWT(userUUID, "secret", -time.Hour)
	tokenWithDifferentIssuer := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJub3QtY2hpcnB5Iiwic3ViIjoiYzIyOGFmOWMtODM1OC00OWJmLTg1NGYtMGQ5MTJmOTNkMTY4IiwiZXhwIjoxOTE0NTc2NTg4LCJpYXQiOjE3NTkwNTY1ODh9.uvnGQ-w_f-iSDSUS1UUsYPujAjRaeDDfTJV-SvfjQuk"

	tests := []struct {
		scenario         string
		tokenString      string
		tokenSecret      string
		expectedUserUUID uuid.UUID
		shouldErr        bool
		expectedError    error
	}{
		{
			scenario:         "accepts unexpired token with valid secret and issued by chirpy",
			tokenString:      validToken,
			tokenSecret:      "secret",
			expectedUserUUID: userUUID,
			shouldErr:        false,
			expectedError:    nil,
		},
		{
			scenario:         "rejects token when secret is incorrect",
			tokenString:      validToken,
			tokenSecret:      "weird-secret",
			expectedUserUUID: uuid.Nil,
			shouldErr:        true,
			expectedError:    jwt.ErrSignatureInvalid,
		},
		{
			scenario:         "rejects incorrectly formatted token",
			tokenString:      "derp.derp.derp",
			tokenSecret:      "secret",
			expectedUserUUID: uuid.Nil,
			shouldErr:        true,
			expectedError:    jwt.ErrTokenMalformed,
		},
		{
			scenario:         "rejects empty token",
			tokenString:      "",
			tokenSecret:      "secret",
			expectedUserUUID: uuid.Nil,
			shouldErr:        true,
			expectedError:    jwt.ErrTokenMalformed,
		},
		{
			scenario:         "rejects expired token",
			tokenString:      expiredToken,
			tokenSecret:      "secret",
			expectedUserUUID: uuid.Nil,
			shouldErr:        true,
			expectedError:    jwt.ErrTokenExpired,
		},
		{
			scenario:         "rejects token that is not issued by chirpy",
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

func TestTokenLifetimeConstraints(t *testing.T) {
	tests := []struct {
		scenario              string
		expiresIn             time.Duration
		expectedTokenLifetime time.Duration
	}{
		{
			scenario:              "allows requested lifetime when within the maximum limit",
			expiresIn:             30 * time.Minute,
			expectedTokenLifetime: 30 * time.Minute,
		},
		{
			scenario:              "caps requested lifetime to the maximum limit",
			expiresIn:             2 * time.Hour,
			expectedTokenLifetime: time.Hour,
		},
		{
			scenario:              "sets lifetime to minimum if it is not provided",
			expiresIn:             0,
			expectedTokenLifetime: time.Hour,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.scenario, func(t *testing.T) {
			actualTokenLifetime := SetTokenLifetime(testCase.expiresIn)
			if actualTokenLifetime != testCase.expectedTokenLifetime {
				t.Errorf("token lifetimes don't match: got [%v] want [%v]", actualTokenLifetime, testCase.expectedTokenLifetime)
				return
			}

		})
	}
}
