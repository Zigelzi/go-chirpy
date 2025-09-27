package main

import "testing"

func TestChirpProfanityFiltering(t *testing.T) {
	testCases := []struct {
		scenario             string
		text                 string
		expectedSensoredText string
	}{
		{
			scenario:             "allows clean chirp to pass through unchanged",
			text:                 "this is example chirp without profanities",
			expectedSensoredText: "this is example chirp without profanities",
		},
		{
			scenario:             "censors lowercase profanity",
			text:                 "this is example chirp with kerfuffle",
			expectedSensoredText: "this is example chirp with ****",
		},
		{
			scenario:             "cencors uppercase profanity",
			text:                 "this is example chirp with KERFUFFLE",
			expectedSensoredText: "this is example chirp with ****",
		},
		{
			scenario:             "cencors capitalized profanity",
			text:                 "this is example chirp with Kerfuffle",
			expectedSensoredText: "this is example chirp with ****",
		},
		{
			scenario:             "cencors multiple profanities in single chirp",
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
		t.Run(testCase.scenario, func(t *testing.T) {
			actualSensoredText := sensorProfanities(testCase.text, unallowedWords)
			if actualSensoredText != testCase.expectedSensoredText {
				t.Errorf("sensored texts don't match: got [%s] want [%s]", actualSensoredText, testCase.expectedSensoredText)
			}
		})
	}
}
