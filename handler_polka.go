package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Zigelzi/go-chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	// Client must have API key that matches POLKA_API_KEY env var
	// Event must be user.upgraded
	// If not, respond with 204
	// Check if user exists
	// Upgrade them to Chirpy red
	// If not, respond with 404
	type polkaWebhookRequestData struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		switch err {
		case auth.ErrNoAuthorizationHeader:
			{
				respondWithError(w, "Authorization header is missing", http.StatusUnauthorized, err)
				return
			}
		case auth.ErrNoAuthorizationType:
			{
				respondWithError(w, "Authorization header is not in ApiKey <credentials> format", http.StatusUnauthorized, err)
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

	if apiKey != cfg.polkaApiKey {
		respondWithError(w, "Invalid API key.", http.StatusUnauthorized, nil)
		return
	}

	decoder := json.NewDecoder(r.Body)
	webhookData := polkaWebhookRequestData{}
	err = decoder.Decode(&webhookData)
	if err != nil {
		respondWithError(w, "Invalid JSON in the request body", http.StatusBadRequest, err)
		return
	}

	if webhookData.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	isExistingUser, err := cfg.db.UserExists(r.Context(), webhookData.Data.UserID)
	if err != nil {
		respondWithError(w, "Something went wrong when fetching user. User was not upgraded to Chirpy Red", http.StatusNotFound, err)
		return
	}

	if isExistingUser == false {
		errorMessage := fmt.Sprintf("User with ID %v was not found", webhookData.Data.UserID)
		respondWithError(w, errorMessage, http.StatusNotFound, nil)
		return
	}

	err = cfg.db.UpgradeUserToChirpyRed(r.Context(), webhookData.Data.UserID)
	if err != nil {
		respondWithError(w, "Something went wrong and user was not upgraded to Chirpy Red", http.StatusNotFound, err)
		return
	}
	fmt.Printf("Upgrading user %v to Chirpy Red\n", webhookData.Data.UserID)
	w.WriteHeader(http.StatusNoContent)
}
