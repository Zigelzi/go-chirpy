package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Zigelzi/go-chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	type getChirpResponse struct {
		Chirp `json:"chirp"`
	}

	pathChirpID := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(pathChirpID)
	if err != nil {
		respondWithError(w, "Parameter must be valid UUID", http.StatusBadRequest, err)
		return
	}
	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, fmt.Sprintf("Chirp with ID %s was not found", chirpID), http.StatusNotFound, err)
			return
		}
		respondWithError(w, "Something went wrong when fetching the chirp", http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK,
		Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		},
	)

}

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	type getAllChirpsResponse struct {
		Chirps []Chirp `json:"chirps"`
	}

	// Service
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, "Failed to get chirps", http.StatusInternalServerError, err)
	}

	allChirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		allChirps = append(allChirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, allChirps)
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
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
	newChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:     uuid.New(),
		Body:   cleanedBody,
		UserID: requestData.UserID,
	})
	if err != nil {
		respondWithError(w, "Something went wrong when creating new chirp", http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, createChirpResponse{
		Chirp: Chirp{
			ID:        newChirp.ID,
			CreatedAt: newChirp.CreatedAt,
			UpdatedAt: newChirp.UpdatedAt,
			Body:      newChirp.Body,
			UserID:    newChirp.UserID,
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
