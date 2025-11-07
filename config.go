package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"github.com/Zigelzi/go-chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	env            string
	jwtSecret      string
	polkaApiKey    string
}

func initConfig(newDB *sql.DB) apiConfig {
	defaultVariables := []string{}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		defaultVariables = append(defaultVariables, "JWT_SECRET")
		jwtSecret = "development-is-hard"
	}
	currentEnv := os.Getenv("ENVIRONMENT")
	if currentEnv == "" {
		defaultVariables = append(defaultVariables, "ENVIRONMENT")
		currentEnv = "development"
	}
	polkaApiKey := os.Getenv("POLKA_API_KEY")
	if polkaApiKey == "" {
		defaultVariables = append(defaultVariables, "POLKA_API_KEY")
		polkaApiKey = "webhooks-are-hard"
	}

	if len(defaultVariables) > 0 {
		log.Printf("Configuration initialized with %d default variables:", len(defaultVariables))
		for _, variable := range defaultVariables {
			log.Println(variable)
		}
	}

	return apiConfig{
		fileServerHits: atomic.Int32{},
		db:             database.New(newDB),
		env:            currentEnv,
		jwtSecret:      jwtSecret,
		polkaApiKey:    polkaApiKey,
	}
}

func initDB() (*sql.DB, error) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DB_URL environment variable is missing")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %s", err)
	}
	return db, nil
}
