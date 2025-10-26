package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Zigelzi/go-chirpy/internal/auth"
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
		Body string `json:"body"`
	}

	type createChirpResponse struct {
		Chirp
	}
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		switch err {
		case auth.ErrNoAuthorizationHeader:
			{
				respondWithError(w, "Authorization header is missing", http.StatusUnauthorized, err)
				return
			}
		case auth.ErrNoAuthorizationType:
			{
				respondWithError(w, "Authorization header is not in Bearer <credentials> format", http.StatusUnauthorized, err)
				return
			}
		case auth.ErrNoAuthorizationCredentials:
			{
				respondWithError(w, "Authorization header is missing creadentials", http.StatusUnauthorized, err)
				return
			}
		default:
			{
				respondWithError(w, "Internal server error when parsing Authorization header", http.StatusInternalServerError, err)
				return
			}
		}
	}
	userUUID, err := auth.ValidateJWT(authToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, "Invalid authorization token", http.StatusUnauthorized, err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	requestData := chirpRequestData{}
	err = decoder.Decode(&requestData)
	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError, err)
		return
	}

	if strings.TrimSpace(requestData.Body) == "" {
		respondWithError(w, "Chirp body can't be empty", http.StatusBadRequest, nil)
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
		UserID: userUUID,
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

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	/*
		User can delete their chirps

		You need to provide the UUID of the chirp to be deleted in the path of the request.

		You can only delete chirps that you are author of.
		You receive unauthorized response, if you try to delete a chirp that you're not author of.
		You get action successful response, when chirp has been deleted successfully.
		You get action successful response, when you delete same chirp multiple times.

		You can get feedback if you don't provide the ChirpID as valid UUID.
		You can't get deleted chirps from the API, when you query for individual chirps.
		You can't get deleted chirps from the API, when you query for all chirps.
		Deleted chirps are still available in DB.
	*/

	pathChirpID := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(pathChirpID)
	if err != nil {
		respondWithError(w, "Parameter must be valid UUID", http.StatusBadRequest, err)
		return
	}

	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		switch err {
		case auth.ErrNoAuthorizationHeader:
			{
				respondWithError(w, "Authorization header is missing", http.StatusUnauthorized, err)
				return
			}
		case auth.ErrNoAuthorizationType:
			{
				respondWithError(w, "Authorization header is not in Bearer <credentials> format", http.StatusUnauthorized, err)
				return
			}
		case auth.ErrNoAuthorizationCredentials:
			{
				respondWithError(w, "Authorization header is missing creadentials", http.StatusUnauthorized, err)
				return
			}
		default:
			{
				respondWithError(w, "Internal server error when parsing Authorization header", http.StatusInternalServerError, err)
				return
			}
		}
	}
	userID, err := auth.ValidateJWT(authToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, "Invalid authorization token", http.StatusUnauthorized, err)
		return
	}

	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorMessage := fmt.Sprintf("Chirp with ID %v was not found", chirpID)
			respondWithError(w, errorMessage, http.StatusNotFound, nil)
			return
		}
		respondWithError(w, "Failed to fetch the chirp with ID "+chirpID.String(), http.StatusInternalServerError, err)
		return
	}

	if dbChirp.DeletedAt.Valid {
		fmt.Printf("User %v attempted to delete chirp %v again. It is already deleted at %v\n", userID, chirpID, dbChirp.DeletedAt.Time)
		respondWithJSON(w, http.StatusNoContent, struct{}{})
		return
	}

	if dbChirp.UserID != userID {
		respondWithError(w, "You can delete only your own chirps", http.StatusForbidden, nil)
		return
	}
	deletedChirp, err := cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil && errors.Is(err, sql.ErrNoRows) == false {
		respondWithError(w, "Failed to delete chirp with ID "+chirpID.String(), http.StatusInternalServerError, err)
		return
	}
	fmt.Printf("User %v deleted chirp %v at %v\n", userID, chirpID, deletedChirp.DeletedAt.Time)
	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
