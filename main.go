package main

import (
	"log"
	"net/http"
)

func main() {
	address := ":8080"
	log.Printf("Starting server on address %s", address)
	mux := http.NewServeMux()
	mux.Handle("/app/",
		http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", handleHealth)
	server := http.Server{
		Handler: mux,
		Addr:    address,
	}
	server.ListenAndServe()
}

func handleHealth(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(http.StatusText(http.StatusOK)))
}
