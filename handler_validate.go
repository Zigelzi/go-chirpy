package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

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

	unallowedWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	cleanedBody := sensorProfanities(chirpToValidate.Body, unallowedWords)

	respondWithJSON(w, http.StatusOK, respVals{
		CleanedBody: cleanedBody,
	})
}

func sensorProfanities(text string, unallowedWords map[string]struct{}) string {
	words := strings.Split(text, " ")
	for wordIndex, word := range words {
		lowercaseWord := strings.ToLower(word)
		if _, ok := unallowedWords[lowercaseWord]; ok {
			words[wordIndex] = "****"
		}
	}
	sensoredText := strings.Join(words, " ")
	return sensoredText
}
