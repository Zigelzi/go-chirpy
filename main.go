package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

func main() {
	address := ":8080"
	cfg := apiConfig{
		fileServerHits: atomic.Int32{},
	}
	log.Printf("Starting server on address %s", address)

	mux := http.NewServeMux()
	mux.Handle("/app/",
		middlewareLogging(
			cfg.middlewareIncrementViews(
				http.StripPrefix("/app/", http.FileServer(http.Dir("."))),
			),
		),
	)
	mux.Handle("GET /api/healthz", middlewareLogging(http.HandlerFunc(handleHealth)))
	mux.Handle("GET /admin/metrics", middlewareLogging(http.HandlerFunc(cfg.handleMetrics)))
	mux.Handle("POST /api/reset", middlewareLogging(http.HandlerFunc(cfg.handleReset)))

	server := http.Server{
		Handler: mux,
		Addr:    address,
	}
	server.ListenAndServe()
}
