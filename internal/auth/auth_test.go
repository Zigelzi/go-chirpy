package auth

import (
	"errors"
	"net/http"
	"testing"
)

func TestParsingAuthenticationToken(t *testing.T) {
	defaultHeaders := http.Header{}
	defaultHeaders.Add("Content-Type", "application/json")
	tests := []struct {
		scenario            string
		headers             http.Header
		expectedTokenString string
		shouldErr           bool
		expectedErr         error
	}{
		{
			scenario: "extracts token from valid authorization header",
			headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer this.is.jwt.token"},
			},
			expectedTokenString: "this.is.jwt.token",
			shouldErr:           false,
			expectedErr:         nil,
		},
		{
			scenario: "extracts token from valid authorization header and trims the whitespaces",
			headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer this.is.jwt.token      "},
			},
			expectedTokenString: "this.is.jwt.token",
			shouldErr:           false,
			expectedErr:         nil,
		},
		{
			scenario: "rejects missing authorization header",
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			expectedTokenString: "",
			shouldErr:           true,
			expectedErr:         ErrNoAuthorizationHeader,
		},
		{
			scenario: "rejects authorization header without token string",
			headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer"},
			},
			expectedTokenString: "",
			shouldErr:           true,
			expectedErr:         ErrNoTokenString,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.scenario, func(t *testing.T) {
			actualTokenString, err := GetBearerToken(testCase.headers)

			if testCase.shouldErr {
				if err == nil {
					t.Errorf("expected error: got [nil]")
					return
				}
				if errors.Is(err, testCase.expectedErr) == false {
					t.Errorf("Errors don't match: got [%v] want [%v]", err, testCase.expectedErr)
					return
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: got [%v] want [nil]", err)
					return
				}
			}

			if actualTokenString != testCase.expectedTokenString {
				t.Errorf("token strings don't match: got [%s] want [%s]", actualTokenString, testCase.expectedTokenString)
				return
			}
		})
	}
}
