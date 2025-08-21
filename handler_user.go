package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Zigelzi/go-chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type requestData struct {
		Email string `json:"email"`
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

	if !isValidEmail(userData.Email) {
		respondWithError(w, "Email not in 'example@domain.com' format", http.StatusBadRequest, nil)
		return
	}

	newUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:    uuid.New(),
		Email: strings.TrimSpace(userData.Email),
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

func isValidEmail(email string) bool {
	if len(email) > 254 { // RFC 5321 limit
		return false
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
