package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type chirpRequestData struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type createChirpResponse struct {
		Chirp
	}

	decoder := json.NewDecoder(r.Body)
	requestData := chirpRequestData{}
	err := decoder.Decode(&requestData)
	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError, err)
		return
	}

	if strings.TrimSpace(requestData.Body) == "" {
		respondWithError(w, "Chirp body can't be empty", http.StatusBadRequest, nil)
		return
	}
	if requestData.UserID == uuid.Nil {
		respondWithError(w, "User ID can't be empty", http.StatusBadRequest, nil)
		return
	}

	// Service
	cleanedBody, err := validateChirp(requestData.Body)
	if err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	respondWithJSON(w, http.StatusCreated, createChirpResponse{
		Chirp: Chirp{
			Body: cleanedBody,
		},
	})
}

func validateChirp(body string) (string, error) {
	const maxLength = 140
	if len(body) > maxLength {
		return "", fmt.Errorf("chirp body is over %d characters (%d)", maxLength, len(body))
	}

	unallowedWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	cleanedBody := sensorProfanities(body, unallowedWords)

	return cleanedBody, nil
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
