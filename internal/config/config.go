package config

import (
	"fmt"
	"os"
)

type Config struct {
	OpenRouterAPIKey string
	AgentLanguage    string
	AgentModel       string
}

func Load() (*Config, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY is not set")
	}

	language := os.Getenv("LANGUAGE")
	if language == "" {
		language = "en"
	}

	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "nvidia/nemotron-3-ultra-550b-a55b:free"
	}

	return &Config{
		OpenRouterAPIKey: apiKey,
		AgentLanguage:    language,
		AgentModel:       model,
	}, nil
}
