package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

var unallowedWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	type respVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	chirpToValidate := chirp{}

	err := decoder.Decode(&chirpToValidate)
	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError, err)
		return
	}
	if chirpToValidate.Body == "" {
		respondWithError(w, "Chirp body is missing", http.StatusBadRequest, nil)
		return
	}
	if len(chirpToValidate.Body) > 140 {
		respondWithError(w, "Chirp is too long (> 140 chars)", http.StatusBadRequest, nil)
		return
	}
	if hasProfanities(chirpToValidate.Body) {
		respondWithJSON(w, http.StatusOK, respVals{
			CleanedBody: sensorProfanities(chirpToValidate.Body),
		})
		return
	}

	respondWithJSON(w, http.StatusOK, respVals{
		CleanedBody: chirpToValidate.Body,
	})
}

func hasProfanities(text string) bool {
	for _, word := range unallowedWords {
		if strings.Contains(strings.ToLower(text), word) {
			return true
		}
	}

	return false
}

func sensorProfanities(text string) string {
	words := strings.Split(text, " ")
	for wordIndex, word := range words {
		for _, profanity := range unallowedWords {
			if strings.ToLower(word) == profanity {
				words[wordIndex] = "****"
			}
		}
	}
	sensoredText := strings.Join(words, " ")
	return sensoredText
}
