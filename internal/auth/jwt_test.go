package auth

import (
	"strings"
	"testing"
	"time"

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

func TestValidateJWT(t *testing.T) {
	userUUID := uuid.New()
	validToken, _ := MakeJWT(userUUID, "secret", time.Hour)

	tests := []struct {
		name             string
		tokenString      string
		tokenSecret      string
		expectedUserUUID uuid.UUID
		wantErr          bool
		expectedError    string
	}{
		{
			name:             "JWT with valid format and secret is valid",
			tokenString:      validToken,
			tokenSecret:      "secret",
			expectedUserUUID: userUUID,
			wantErr:          false,
			expectedError:    "",
		},
		{
			name:             "JWT with incorrect secret is invalid ",
			tokenString:      validToken,
			tokenSecret:      "weird-secret",
			expectedUserUUID: uuid.Nil,
			wantErr:          true,
			expectedError:    "signature is invalid",
		},
		{
			name:             "JWT with incorrect format is invalid ",
			tokenString:      "derp.derp.derp",
			tokenSecret:      "secret",
			expectedUserUUID: uuid.Nil,
			wantErr:          true,
			expectedError:    "token is malformed",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			actualUserUUID, err := ValidateJWT(testCase.tokenString, testCase.tokenSecret)
			if testCase.wantErr {
				if err == nil {
					t.Errorf("Expected error but got [nil]")
					return
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
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
