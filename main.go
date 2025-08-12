package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareIncrementViews(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileServerHits.Load())))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits = atomic.Int32{}
}

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
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("GET /metrics", cfg.handleMetrics)
	mux.HandleFunc("POST /reset", cfg.handleReset)
	server := http.Server{
		Handler: mux,
		Addr:    address,
	}
	server.ListenAndServe()
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func middlewareLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s request to %s", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}
