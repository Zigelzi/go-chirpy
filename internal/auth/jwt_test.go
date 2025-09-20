package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateJWT(t *testing.T) {
	t.Run("JWT expiring in the future and issued by chirpy is valid", func(t *testing.T) {
		tokenSecret := "testing-is-fun"
		expectedUserUUID := uuid.New()

		token, _ := MakeJWT(expectedUserUUID, tokenSecret, time.Hour)
		actualUserUUID, err := ValidateJWT(token, tokenSecret)
		if actualUserUUID != expectedUserUUID {
			t.Errorf("User UUIDs don't match: got [%v] want [%v]", actualUserUUID, expectedUserUUID)
		}
		if err != nil {
			t.Errorf("Got unexpected error when validating the JWT: %v", err)
		}
	})
	t.Run("JWT not issued by chirpy is invalid", func(t *testing.T) {
		tokenSecret := "testing-is-fun"
		expectedUserUUID := uuid.UUID{}
		expectedError := "token is not issued by " + issuerName
		// Expires on 2030-09-19 14:48:48.406655373 +0000 UTC, generate new before that.
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJpbmNvcnJlY3QtaXNzdWVyIiwic3ViIjoiNjAyYTY0YTgtYmJiMi00MTc4LWFiYmItNjcxOWQzZmRjODgwIiwiZXhwIjoxOTE2MDU5NzI4LCJpYXQiOjE3NTgzNzk3Mjh9.nAq3WD2ay5klCfZcgOuWQ5rOdBhTYko07saFO7rBKsQ"
		actualUserUUID, err := ValidateJWT(token, tokenSecret)
		if actualUserUUID != expectedUserUUID {
			t.Errorf("User UUIDs don't match: got [%v] want [%v]", actualUserUUID, expectedUserUUID)
		}
		if err == nil {
			t.Errorf("Expected error but got [nil]")
			return
		}
		if strings.Contains(err.Error(), expectedError) == false {
			t.Errorf("Errors don't match: got [%s] want [%s]", err, expectedError)
		}
	})
	t.Run("JWT expiring in past is invalid", func(t *testing.T) {
		tokenSecret := "testing-is-fun"
		expectedUserUUID := uuid.UUID{}
		expectedError := "token is expired"
		token, _ := MakeJWT(expectedUserUUID, tokenSecret, -time.Hour)
		actualUserUUID, err := ValidateJWT(token, tokenSecret)
		if actualUserUUID != expectedUserUUID {
			t.Errorf("User UUIDs don't match: got [%v] want [%v]", actualUserUUID, expectedUserUUID)
		}
		if err == nil {
			t.Errorf("Expected error but got [nil]")
			return
		}
		if strings.Contains(err.Error(), expectedError) == false {
			t.Errorf("Errors don't match: got [%s] want [%s]", err, expectedError)
		}
	})
	t.Run("JWT signed with different secret is invalid", func(t *testing.T) {
		tokenSecret := "testing-is-fun"
		expectedUserUUID := uuid.UUID{}
		expectedError := "token signature is invalid"
		token, _ := MakeJWT(expectedUserUUID, "not-valid-secret", -time.Hour)
		actualUserUUID, err := ValidateJWT(token, tokenSecret)
		if actualUserUUID != expectedUserUUID {
			t.Errorf("User UUIDs don't match: got [%v] want [%v]", actualUserUUID, expectedUserUUID)
		}
		if err == nil {
			t.Errorf("Expected error but got [nil]")
			return
		}
		if strings.Contains(err.Error(), expectedError) == false {
			t.Errorf("Errors don't match: got [%s] want [%s]", err, expectedError)
		}
	})
}
