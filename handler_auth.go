package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Zigelzi/go-chirpy/internal/auth"
	"github.com/Zigelzi/go-chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type loginRequestData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type loginReponse struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	loginData := loginRequestData{}
	err := decoder.Decode(&loginData)
	if err != nil {
		respondWithError(w, "Request body must be valid JSON ", http.StatusBadRequest, err)
		return
	}

	dbUser, err := cfg.db.GetUserByEmail(r.Context(), loginData.Email)
	if err != nil {
		respondWithError(w, "Incorrect email or password", http.StatusUnauthorized, err)
		return
	}

	// Service starts here?
	err = auth.CheckHashedPassword(loginData.Password, dbUser.HashedPassword)
	if err != nil {
		respondWithError(w, "Incorrect email or password", http.StatusUnauthorized, err)
		return
	}

	accessToken, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, "Unexpected error when logging in the user", http.StatusInternalServerError, err)
		return
	}

	/*
		Story: User can log out from the system
		---

		You can stay logged in for 60 days on same client without needing to login again.
		You can stay logged in log in from multiple clients.
		You can get feedback in the API if something goes wrong when creating the refresh token
		in login.
		You can log out from the service.
		You can't access the authorized parts of the systems with the same client after logging out.

		Task: Add creating and returning a refresh token for an user when they log in
		Task: Add endpoint to refresh the access token with refresh token
		Task: Add endpoint to revoke the refresh token
	*/

	/*
		Task: Create new refresh token for an user when they log in and return it in the response.
		---
		You can get refresh token in the response, when you log in.
		Refresh token is stored in the database.
		Refresh tokens always expire in 60 days.
		Access tokens always expire in 1 hour.
		User can have multiple active refresh tokens.
		Refresh tokens always belong to only one user.
		Refresh tokens get deleted from database when the user is deleted.
		You can get feedback in the API if something goes wrong when creating the refresh token
		in login.
	*/
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, "Unexpected error when logging in the user", http.StatusInternalServerError, err)
		return
	}

	// Possibly separate function in auth which responsibility is to store the refresh token.
	refreshTokenLifetime := time.Hour * 24 * 60
	err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(refreshTokenLifetime),
		UserID:    dbUser.ID,
	})

	if err != nil {
		respondWithError(w, "Unexpected error when logging in the user", http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, loginReponse{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handleRefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	/*
		Task: Create new endpoint to refresh the access token.
		--

		You need to provide refresh token in the Authorization request header.
		You can get new access token in the response body, when the refresh token is valid
		Valid refresh token exists in the DB, is not expired and is not revoked.
		You can get feedback what is wrong, if you don't provide refresh token in the header or
		it's not in expected format or without expected content.
		You get unauthorized response, when the refresh token is invalid.
	*/
	type refreshAccessTokenResponse struct {
		Token string `json:"token"`
	}
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, "Unexpected error when refreshing the access token", http.StatusInternalServerError, err)
		return
	}
	log.Printf("Refreshing access token with refresh token: %s", refreshToken)
	userId, err := cfg.db.GetUserFromValidRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, "", http.StatusUnauthorized, err)
			return
		}
		respondWithError(w, "Unexpected error when refreshing the access token", http.StatusInternalServerError, err)
		return
	}
	accessToken, err := auth.MakeJWT(userId, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, "Unexpected error when refreshing the access token", http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, refreshAccessTokenResponse{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handleRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	/*
		Task: Add endpoint to revoke the refresh token
		---
		You need to provide refresh token in the Authorization header.
		Active refresh token is revoked, if it exists in database.
		You can't use revoked refresh token to generate new access tokens.
		You can get feedback what is wrong, if you don't provide refresh token in the header or
		it's not in expected format or without expected content.
	*/

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, "Unexpected error when revoking the refresh token", http.StatusInternalServerError, err)
		return
	}
	log.Printf("Revoking refresh token: %s", refreshToken)
	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, "Unexpected error when revoking the refresh token", http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
