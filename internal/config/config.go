package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	APIKey     string
	Compiler   string
	CppStd     string
	Debug      bool
	CacheDir   string
	DocsDir    string
	UseCache   bool
	MaxRetries int
}

func Load() (*Config, error) {
	// Find the flow installation directory
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}
	baseDir := filepath.Dir(filepath.Dir(execPath))

	// Try to find docs relative to executable, fallback to /opt/flow
	docsDir := filepath.Join(baseDir, "docs")
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		docsDir = "/opt/flow/docs"
	}

	cacheDir := filepath.Join(baseDir, "cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		cacheDir = "/opt/flow/cache"
	}

	cfg := &Config{
		APIKey:     getEnv("ANTHROPIC_API_KEY", ""),
		Compiler:   getEnv("FLOW_COMPILER", "g++"),
		CppStd:     getEnv("FLOW_STD", "c++17"),
		Debug:      getEnv("FLOW_DEBUG", "false") == "true",
		CacheDir:   cacheDir,
		DocsDir:    docsDir,
		UseCache:   getEnv("FLOW_CACHE", "true") == "true",
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
