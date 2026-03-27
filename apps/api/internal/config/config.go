package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL       string
	FREDAPIKey        string
	Port              string
	InternalAuthToken string
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	fredKey := os.Getenv("FRED_API_KEY")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	authToken := os.Getenv("INTERNAL_AUTH_TOKEN")

	return &Config{
		DatabaseURL:       dbURL,
		FREDAPIKey:        fredKey,
		Port:              port,
		InternalAuthToken: authToken,
	}, nil
}
