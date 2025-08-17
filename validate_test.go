package main

import "testing"

func TestSensorProfanities(t *testing.T) {
	testCases := []struct {
		name                 string
		text                 string
		expectedSensoredText string
	}{
		{
			name:                 "sentence without profanities isn't sensored",
			text:                 "this is example chirp without profanities",
			expectedSensoredText: "this is example chirp without profanities",
		},
		{
			name:                 "sentence with single lowercase profanity is sensored",
			text:                 "this is example chirp with kerfuffle",
			expectedSensoredText: "this is example chirp with ****",
		},
		{
			name:                 "sentence with single uppercase profanity is sensored",
			text:                 "this is example chirp with KERFUFFLE",
			expectedSensoredText: "this is example chirp with ****",
		},
		{
			name:                 "sentence with single capitalized profanity is sensored",
			text:                 "this is example chirp with Kerfuffle",
			expectedSensoredText: "this is example chirp with ****",
		},
		{
			name:                 "sentence with two different profanities is sensored",
			text:                 "this is example chirp with Kerfuffle and sharbert",
			expectedSensoredText: "this is example chirp with **** and ****",
		},
	}

	unallowedWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualSensoredText := sensorProfanities(testCase.text, unallowedWords)
			if actualSensoredText != testCase.expectedSensoredText {
				t.Errorf("sensored texts don't match: got [%s] want [%s]", actualSensoredText, testCase.expectedSensoredText)
			}
		})
	}
}
