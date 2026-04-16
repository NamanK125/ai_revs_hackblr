// Package config loads application configuration from environment variables.
package config

import (
	"errors"
	"os"
	"strings"
)

// Config holds all application settings loaded from environment.
type Config struct {
	// Vapi credentials
	VapiPublicKey    string
	VapiPrivateKey   string
	VapiSharedSecret string
	VapiAssistantID  string
	VapiServerURL    string

	// Qdrant vector database
	QdrantURL string

	// Python FastEmbed sidecar
	EmbedderURL string

	// Server port
	Port string

	// GitHub seed repositories for corpus ingestion
	GitHubSeedRepos []string
}

// Load reads configuration from environment variables.
// Returns error if required fields (VapiPublicKey, VapiPrivateKey) are empty.
func Load() (*Config, error) {
	cfg := &Config{
		VapiPublicKey:    os.Getenv("VAPI_PUBLIC_KEY"),
		VapiPrivateKey:   os.Getenv("VAPI_PRIVATE_KEY"),
		VapiSharedSecret: os.Getenv("VAPI_SHARED_SECRET"),
		VapiAssistantID:  os.Getenv("VAPI_ASSISTANT_ID"),
		VapiServerURL:    os.Getenv("VAPI_SERVER_URL"),
		EmbedderURL:      os.Getenv("EMBEDDER_URL"),
		Port:             os.Getenv("PORT"),
	}

	// Set defaults
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.QdrantURL == "" {
		cfg.QdrantURL = os.Getenv("QDRANT_URL")
		if cfg.QdrantURL == "" {
			cfg.QdrantURL = "http://localhost:6334"
		}
	} else {
		cfg.QdrantURL = os.Getenv("QDRANT_URL")
		if cfg.QdrantURL == "" {
			cfg.QdrantURL = "http://localhost:6334"
		}
	}
	if cfg.EmbedderURL == "" {
		cfg.EmbedderURL = "http://localhost:8001"
	}

	// Parse GitHub seed repos from comma-separated env var
	repos := os.Getenv("GITHUB_SEED_REPOS")
	if repos != "" {
		parts := strings.Split(repos, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				cfg.GitHubSeedRepos = append(cfg.GitHubSeedRepos, trimmed)
			}
		}
	}

	// Validate required fields
	if cfg.VapiPublicKey == "" {
		return nil, errors.New("VAPI_PUBLIC_KEY is required")
	}
	if cfg.VapiPrivateKey == "" {
		return nil, errors.New("VAPI_PRIVATE_KEY is required")
	}

	return cfg, nil
}