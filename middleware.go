package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) middlewareIncrementViews(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func middlewareLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s request to %s", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}
