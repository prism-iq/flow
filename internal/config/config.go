package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	LogLevel       string
	LLMServiceURL  string
	RAGServiceURL  string
	RustServiceURL string
	AllowedOrigins []string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:           getEnv("PORT", "8080"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		LLMServiceURL:  getEnv("LLM_SERVICE_URL", "http://localhost:8001"),
		RAGServiceURL:  getEnv("RAG_SERVICE_URL", "http://localhost:8002"),
		RustServiceURL: getEnv("RUST_SERVICE_URL", "http://localhost:8003"),
		AllowedOrigins: []string{"*"},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
