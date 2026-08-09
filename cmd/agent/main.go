package main

import (
	"fmt"
	"os"

	"aider/internal/agent"
	"aider/internal/cli"
	"aider/internal/config"
	"aider/internal/executor"
	"aider/internal/llm"
	"aider/internal/security"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	config.Load()

	// Здесь оставляем твою существующую
	// загрузку конфигурации.

	llmClient := llm.New(
		config.apiKey,
		config.model,
		config.language,
	)

	ag := agent.New(llmClient)

	exec := executor.New()

	policy := security.DefaultPolicy()

	app := cli.New(
		&ag,
		exec,
		policy,
	)

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(
			os.Stderr,
			"error:",
			err,
		)

		os.Exit(1)
	}
}
