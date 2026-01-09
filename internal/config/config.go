package config

import (
	"os"
)

type Config struct {
	Compiler   string
	CppStd     string
	Debug      bool
	MaxRetries int
}

func Load() (*Config, error) {
	cfg := &Config{
		Compiler:   getEnv("FLOW_COMPILER", "g++"),
		CppStd:     getEnv("FLOW_STD", "c++20"),
		Debug:      getEnv("FLOW_DEBUG", "false") == "true",
		MaxRetries: 3,
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
