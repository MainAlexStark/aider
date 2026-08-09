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

	if len(os.Args) >= 2 &&
		(os.Args[1] == "-h" || os.Args[1] == "--help") {

		cli.PrintHelp()
		return
	}

	_ = godotenv.Load()
	cnf, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Здесь оставляем твою существующую
	// загрузку конфигурации.

	llmClient := llm.New(
		cnf.OpenRouterAPIKey,
		cnf.Model,
		cnf.Language,
	)

	exec := executor.New()

	policy := security.DefaultPolicy()

	ag := agent.New(llmClient, exec, policy)

	app := cli.New(
		ag,
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
