package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Zigelzi/go-chirpy/internal/auth"
	"github.com/Zigelzi/go-chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type requestData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type createUserReponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	userData := requestData{}
	err := decoder.Decode(&userData)
	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError, err)
		return
	}

	if strings.TrimSpace(userData.Email) == "" {
		respondWithError(w, "Email is required field", http.StatusBadRequest, nil)
		return
	}

	if isInvalidEmail(userData.Email) {
		respondWithError(w, "Email not in 'example@domain.com' format", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(userData.Password) == "" {
		respondWithError(w, "Password is required field", http.StatusBadRequest, nil)
		return
	}

	if isWeakPassword(userData.Password) {
		respondWithError(w, "Password is too weak", http.StatusBadRequest, nil)
		return
	}

	hashedPassword, err := auth.HashPassword(userData.Password)
	if err != nil {
		respondWithError(w, "Something went wrong creating an user", http.StatusInternalServerError, err)
		return
	}

	newUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:             uuid.New(),
		Email:          strings.TrimSpace(userData.Email),
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, "Something went wrong and user wasn't created", http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, createUserReponse{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	})
}

func isInvalidEmail(email string) bool {
	if len(email) > 254 { // RFC 5321 limit
		return true
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return !emailRegex.MatchString(email)
}

func isWeakPassword(password string) bool {
	const minLength = 4
	if len(password) < minLength {
		return true
	}
	return false
}

func (cfg *apiConfig) handleUpdateUserCredentials(w http.ResponseWriter, r *http.Request) {
	/*
		User can update their login credentials
		---
		x You need to provide valid access token in the Authorization header.
		x You get error about unauthorized use, if you don't provide valid access token
		x You need to provide new email and password to be changed.
		x You can get feedback if you don't provide new email and password.

		x You need to provide email in valid email format
		x You can get feedback when you provide email in invalid format.

		x You need to provide password that meets the system password stength requirements.
		x You can get feedback about insufficient password strenth, when you submit password that
		is too weak.

		x You can only change your own credentials, so user UUID contained in the access token.

		x You can only use the new email and password to login, when you successfully change
		your credentials.
		x You can get the the updated user details without password in the response, when you update
		the credentials successfully.

	*/
	type changeCredentialsRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	accessToken, err := auth.GetBearerToken(r.Header)
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
	userId, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, "Invalid access token", http.StatusUnauthorized, err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	credentialsData := changeCredentialsRequest{}
	err = decoder.Decode(&credentialsData)
	if err != nil {
		respondWithError(w, "Invalid JSON in the request body", http.StatusBadRequest, err)
		return
	}

	if strings.TrimSpace(credentialsData.Email) == "" {
		respondWithError(w, "Email is required field", http.StatusBadRequest, nil)
		return
	}

	if isInvalidEmail(credentialsData.Email) {
		respondWithError(w, "Email not in 'example@domain.com' format", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(credentialsData.Password) == "" {
		respondWithError(w, "Password is required field", http.StatusBadRequest, nil)
		return
	}

	if isWeakPassword(credentialsData.Password) {
		respondWithError(w, "Password is too weak", http.StatusBadRequest, nil)
		return
	}

	newHashedPassword, err := auth.HashPassword(credentialsData.Password)
	if err != nil {
		respondWithError(w, "Internal server error. Credentials weren't updated.", http.StatusInternalServerError, err)
		return
	}

	updatedCredentials, err := cfg.db.UpdateUserCredentials(r.Context(), database.UpdateUserCredentialsParams{
		ID:             userId,
		Email:          credentialsData.Email,
		HashedPassword: newHashedPassword,
	})

	if err != nil {
		respondWithError(w, "Internal server error. Credentials weren't updated.", http.StatusInternalServerError, err)
		return
	}

	type changeCredentialsResponse struct {
		User
	}
	respondWithJSON(w, http.StatusOK, changeCredentialsResponse{
		User: User{
			ID:        updatedCredentials.ID,
			UpdatedAt: updatedCredentials.UpdatedAt,
			CreatedAt: updatedCredentials.CreatedAt,
			Email:     updatedCredentials.Email,
		},
	})
}
