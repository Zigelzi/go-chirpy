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

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type loginRequestData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type loginReponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	userData := loginRequestData{}
	err := decoder.Decode(&userData)
	if err != nil {
		respondWithError(w, "Something went wrong when trying to login the user", http.StatusInternalServerError, err)
		return
	}
	dbUser, err := cfg.db.GetUserByEmail(r.Context(), userData.Email)
	if err != nil {
		respondWithError(w, "Incorrect email or password", http.StatusUnauthorized, err)
		return
	}
	err = auth.CheckHashedPassword(userData.Password, dbUser.HashedPassword)
	if err != nil {
		respondWithError(w, "Incorrect email or password", http.StatusUnauthorized, err)
		return
	}
	respondWithJSON(w, http.StatusOK, loginReponse{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	})
}
