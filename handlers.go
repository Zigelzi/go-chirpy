package main

import (
	"encoding/json"
	"fmt"
	"log"
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
		Error string `json:"error"`
		Valid bool   `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	chirpToValidate := chirp{}
	respBody := respVals{}
	err := decoder.Decode(&chirpToValidate)

	if err != nil {
		log.Printf("error decoding chirp: %v", err)
		respBody.Error = "Something went wrong"
		data, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("error marshaling response: %v", err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(data)
		return
	}
	if chirpToValidate.Body == "" {
		respBody.Valid = false
		respBody.Error = "Chirp body is missing"
		data, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("error marshaling response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return
	}
	if len(chirpToValidate.Body) > 140 {
		respBody.Valid = false
		respBody.Error = "Chirp is too long"
		data, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("error marshaling response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return
	}

	respBody.Valid = true
	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("error marshaling response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
