package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenRouterAPIKey string
	Language         string
	Model            string
	MaxIterations    int
}

func Path() (string, error) {
	home_path, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(
		home_path,
		".config",
		"aider",
		"config.env",
	), nil
}

func Load() (Config, error) {
	configPath, err := Path()
	if err != nil {
		return Config{}, err
	}

	_ = godotenv.Load(configPath)

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	language := os.Getenv("AGENT_LANGUAGE")
	model := os.Getenv("AGENT_MODEL")
	maxIterations, err := strconv.Atoi(os.Getenv("MAX_ITERATIONS"))
	if err != nil {
		maxIterations = 5
	}

	if apiKey == "" {
		return Config{}, fmt.Errorf(
			"OPENROUTER_API_KEY is not set",
		)
	}
	if language == "" {
		language = "en"
	}
	if model == "" {
		model = "nvidia/nemotron-3-ultra-550b-a55b:free"
	}
	if maxIterations == 0 {
		maxIterations = 5
	}

	return Config{
		OpenRouterAPIKey: apiKey,
		Language:         language,
		Model:            model,
		MaxIterations:    maxIterations,
	}, nil
}
