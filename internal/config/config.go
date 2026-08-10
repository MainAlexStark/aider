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
			"OPENROUTER_API_KEY is not set\n Use:\n  aider config set api-key <your_api_key>\n",
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

	// Получаем текущие данные
	if _, err := os.Stat(configPath); err == nil {
		existing, err := godotenv.Read(configPath)
		if err != nil {
			return err
		}

		data = existing
	}

	// Устанавливаем новое значение
	data[key] = value
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Выводим конфиг
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
