package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	address := ":8080"
	godotenv.Load()

	db, err := initDB()
	if err != nil {
		log.Fatalf("initializing database failed: %v", err)
	}
	cfg := initConfig(db)

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

	// External services
	mux.Handle("POST /api/polka/webhooks", middlewareLogging(http.HandlerFunc(cfg.handlePolkaWebhook)))

	// Admin routes
	mux.Handle("GET /admin/metrics", middlewareLogging(http.HandlerFunc(cfg.handleMetrics)))
	mux.Handle("POST /admin/reset", middlewareLogging(http.HandlerFunc(cfg.handleReset)))

	server := http.Server{
		Handler: mux,
		Addr:    address,
	}

	log.Printf("Starting server on address %s for environment [%s]", address, cfg.env)
	server.ListenAndServe()
}
