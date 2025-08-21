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
}

func main() {
	address := ":8080"
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %s", err)
	}
	cfg := apiConfig{
		fileServerHits: atomic.Int32{},
		db:             database.New(db),
		env:            os.Getenv("ENVIRONMENT"),
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

	// API routes
	mux.Handle("GET /api/healthz", middlewareLogging(http.HandlerFunc(handleHealth)))

	// Chirps
	mux.Handle("POST /api/chirps/", middlewareLogging(http.HandlerFunc(handleCreateChirp)))

	// Users
	mux.Handle("POST /api/users", middlewareLogging(http.HandlerFunc(cfg.handleCreateUser)))

	// Admin routes
	mux.Handle("GET /admin/metrics", middlewareLogging(http.HandlerFunc(cfg.handleMetrics)))
	mux.Handle("POST /admin/reset", middlewareLogging(http.HandlerFunc(cfg.handleReset)))

	server := http.Server{
		Handler: mux,
		Addr:    address,
	}
	server.ListenAndServe()
}
