package configs

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	APIURL string // http://localhost:8000/api/v1
	AuthToken string //JWT for current user
	LogLevel string
	Transport string // stdio(for now) or http
}

func LoadConfig() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		authToken = os.Getenv("JWT_TOKEN")
	}

	return &Config{
		APIURL: getEnv("API_URL", "http://localhost:8080/api/v1"),
		AuthToken: authToken,
		LogLevel: getEnv("LOG_LEVEL", "debug"),
		Transport: getEnv("TRANSPORT", "stdio"), // Options: stdio, http
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}