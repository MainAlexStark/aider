package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenRouterAPIKey string
	Language         string
	Model            string
	MaxIterations    string
}

func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(
		home,
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

	// Загружаем конфигурационный файл,
	// если он существует.
	_ = godotenv.Load(configPath)

	apiKey := os.Getenv("OPENROUTER_API_KEY")

	if apiKey == "" {
		return Config{}, fmt.Errorf(
			"OPENROUTER_API_KEY is not set",
		)
	}

	language := os.Getenv("AGENT_LANGUAGE")

	if language == "" {
		language = "ru"
	}

	model := os.Getenv("AGENT_MODEL")

	if model == "" {
		model = "nvidia/nemotron-3-ultra-550b-a55b:free"
	}

	max_iterations := os.Getenv("MAX_ITERATIONS")

	if max_iterations == "" {
		max_iterations = "5"
	}

	return Config{
		OpenRouterAPIKey: apiKey,
		Language:         language,
		Model:            model,
		MaxIterations:    max_iterations,
	}, nil
}

func Set(
	key string,
	value string,
) error {

	configPath, err := Path()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(
		configDir,
		0700,
	); err != nil {
		return err
	}

	data := map[string]string{}

	if _, err := os.Stat(configPath); err == nil {
		existing, err := godotenv.Read(configPath)
		if err != nil {
			return err
		}

		data = existing
	}

	data[key] = value

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}

	defer file.Close()

	for k, v := range data {
		if _, err := fmt.Fprintf(
			file,
			"%s=%s\n",
			k,
			v,
		); err != nil {
			return err
		}
	}

	// Только владелец может читать конфиг.
	return os.Chmod(
		configPath,
		0600,
	)
}
