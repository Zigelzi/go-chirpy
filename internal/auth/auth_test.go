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
			scenario: "rejects authorization header without credentials",
			headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer"},
			},
			expectedTokenString: "",
			shouldErr:           true,
			expectedErr:         ErrNoAuthorizationCredentials,
		},
		{
			scenario: "rejects authorization header without type",
			headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"this.is.jwt.token"},
			},
			expectedTokenString: "",
			shouldErr:           true,
			expectedErr:         ErrNoAuthorizationType,
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

func TestParsingAPIToken(t *testing.T) {
	defaultHeaders := http.Header{}
	defaultHeaders.Add("Content-Type", "application/json")
	tests := []struct {
		scenario         string
		headers          http.Header
		expectedAPIToken string
		shouldErr        bool
		expectedErr      error
	}{
		{
			scenario: "extracts API token from valid authorization header",
			headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"ApiKey polka-api-key"},
			},
			expectedAPIToken: "polka-api-key",
			shouldErr:        false,
			expectedErr:      nil,
		},
		{
			scenario: "extracts API token from valid authorization header and trims the whitespaces",
			headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"ApiKey polka-api-key      "},
			},
			expectedAPIToken: "polka-api-key",
			shouldErr:        false,
			expectedErr:      nil,
		},
		{
			scenario: "rejects missing authorization header",
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			expectedAPIToken: "",
			shouldErr:        true,
			expectedErr:      ErrNoAuthorizationHeader,
		},
		{
			scenario: "rejects authorization header without credentials",
			headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"ApiKey"},
			},
			expectedAPIToken: "",
			shouldErr:        true,
			expectedErr:      ErrNoAuthorizationCredentials,
		},
		{
			scenario: "rejects authorization header without type",
			headers: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"polka-api-key"},
			},
			expectedAPIToken: "",
			shouldErr:        true,
			expectedErr:      ErrNoAuthorizationType,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.scenario, func(t *testing.T) {
			actualTokenString, err := GetAPIKey(testCase.headers)

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

			if actualTokenString != testCase.expectedAPIToken {
				t.Errorf("token strings don't match: got [%s] want [%s]", actualTokenString, testCase.expectedAPIToken)
				return
			}
		})
	}
}

func TestAuthenticationRefreshTokenGeneration(t *testing.T) {
	refreshToken1, _ := MakeRefreshToken()
	refreshToken2, _ := MakeRefreshToken()
	const expectedTokenLength = 64

	if refreshToken1 == refreshToken2 {
		t.Errorf("two identical tokens were created: token 1 [%s], token 2 [%s]", refreshToken1, refreshToken2)
		return
	}

	if len(refreshToken1) != expectedTokenLength {
		t.Errorf("token is incorrect length: got [%d] want [%d]", len(refreshToken1), expectedTokenLength)
		return
	}
}
