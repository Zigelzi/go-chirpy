package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Zigelzi/go-chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	env            string
	jwtSecret      string
}

func main() {
	address := ":8080"
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %s", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "development-is-hard"
	}
	cfg := apiConfig{
		fileServerHits: atomic.Int32{},
		db:             database.New(db),
		env:            os.Getenv("ENVIRONMENT"),
		jwtSecret:      jwtSecret,
	}
	log.Printf("Starting server on address %s for environment [%s]", address, cfg.env)

	mux := http.NewServeMux()

	// Fileserver ("Frontend")
	mux.Handle("/app/",
		middlewareLogging(
			cfg.middlewareIncrementViews(
				http.StripPrefix("/app/", http.FileServer(http.Dir("."))),
			),
		),
	)

	// Health check
	mux.Handle("GET /api/healthz", middlewareLogging(http.HandlerFunc(handleHealth)))

	// Chirps
	mux.Handle("POST /api/chirps", middlewareLogging(http.HandlerFunc(cfg.handleCreateChirp)))
	mux.Handle("GET /api/chirps", middlewareLogging(http.HandlerFunc(cfg.handleGetAllChirps)))
	mux.Handle("GET /api/chirps/{chirpID}", middlewareLogging(http.HandlerFunc(cfg.handleGetChirp)))
	mux.Handle("DELETE /api/chirps/{chirpID}", middlewareLogging(http.HandlerFunc(cfg.handleDeleteChirp)))

	// Users
	mux.Handle("POST /api/users", middlewareLogging(http.HandlerFunc(cfg.handleCreateUser)))
	mux.Handle("PUT /api/users", middlewareLogging(http.HandlerFunc(cfg.handleUpdateUserCredentials)))

	// Authentication
	mux.Handle("POST /api/login", middlewareLogging(http.HandlerFunc(cfg.handleLogin)))
	mux.Handle("POST /api/refresh", middlewareLogging(http.HandlerFunc(cfg.handleRefreshAccessToken)))
	mux.Handle("POST /api/revoke", middlewareLogging(http.HandlerFunc(cfg.handleRevokeRefreshToken)))

	// Admin routes
	mux.Handle("GET /admin/metrics", middlewareLogging(http.HandlerFunc(cfg.handleMetrics)))
	mux.Handle("POST /admin/reset", middlewareLogging(http.HandlerFunc(cfg.handleReset)))

	server := http.Server{
		Handler: mux,
		Addr:    address,
	}
	server.ListenAndServe()
}
