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

	// User gets refresh token to reduce the number of logins in the future.
	// Refresh token is stored in the server and always belongs to one user
	// One user can have only one active refresh token at once.
	// Refresh token always expires after 60 days.
	// Refresh token can be revoked.

	numberOfRevokedTokens, err := cfg.db.RevokeRefreshTokenByUserId(r.Context(), dbUser.ID)
	if err != nil {
		respondWithError(w, "Unexpected error when logging in the user", http.StatusInternalServerError, err)
		return
	}

	if numberOfRevokedTokens > 0 {
		log.Printf("Revoked %d existing token(s) from user %s", numberOfRevokedTokens, dbUser.Email)
	}

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
		User can get new access token with valid refresh token.
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
		User can revoke the refresh token (log out from all devices).

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
