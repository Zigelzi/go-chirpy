package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

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
	type resetResponse struct {
		DeletedCount int64 `json:"deleted_count"`
	}
	if cfg.env != "development" {
		respondWithError(w, "Resetting only allowed in development environment", http.StatusForbidden, nil)
		return
	}

	// Service
	cfg.fileServerHits = atomic.Int32{}
	countOfDeletedUsers, err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		respondWithError(w, "Failed to reset users", http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, resetResponse{
		DeletedCount: countOfDeletedUsers,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
