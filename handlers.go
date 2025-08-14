package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
	`, cfg.fileServerHits.Load())))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits = atomic.Int32{}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	type respVals struct {
		Valid bool `json:"valid"`
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

	respondWithJSON(w, http.StatusOK, respVals{
		Valid: true,
	})
}
