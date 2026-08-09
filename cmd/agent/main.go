package main

import (
	"fmt"
	"os"

	"aider/internal/agent"
	"aider/internal/cli"
	"aider/internal/executor"
	"aider/internal/llm"
	"aider/internal/security"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Здесь оставляем твою существующую
	// загрузку конфигурации.

	llmClient := llm.New(
		os.Getenv("OPENROUTER_API_KEY"),
		os.Getenv("OPENROUTER_MODEL"),
		os.Getenv("LANGUAGE"),
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
