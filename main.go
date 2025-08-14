package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func main() {
	address := ":8080"
	cfg := apiConfig{
		fileServerHits: atomic.Int32{},
	}
	log.Printf("Starting server on address %s", address)

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
	mux.Handle("POST /api/validate_chirp", middlewareLogging(http.HandlerFunc(handleValidateChirp)))

	// Admin routes
	mux.Handle("GET /admin/metrics", middlewareLogging(http.HandlerFunc(cfg.handleMetrics)))
	mux.Handle("POST /admin/reset", middlewareLogging(http.HandlerFunc(cfg.handleReset)))

	server := http.Server{
		Handler: mux,
		Addr:    address,
	}
	server.ListenAndServe()
}
