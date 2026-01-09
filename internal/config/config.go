package config

import (
	"os"

	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

type Config struct {
	Port           string
	LogLevel       string
	LLMServiceURL  string
	RAGServiceURL  string
	RustServiceURL string
	AllowedOrigins []string
	Database       DatabaseConfig
	GraphName      string
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
		Database: DatabaseConfig{
			Host:     getEnv("PG_HOST", "localhost"),
			Port:     getEnv("PG_PORT", "5432"),
			Name:     getEnv("PG_DATABASE", "flow"),
			User:     getEnv("PG_USER", "flow"),
			Password: getEnv("PG_PASSWORD", "flow"),
		},
		GraphName: getEnv("GRAPH_NAME", "flow_graph"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
