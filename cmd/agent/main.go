package main

import (
	"fmt"
	"os"

	"aider/internal/agent"
	"aider/internal/cli"
	"aider/internal/executor"
	"aider/internal/llm"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Println("warning: .env not found")
	}

	llmClient := llm.New(
		os.Getenv("OPENROUTER_API_KEY"),
	)

	executor := executor.New()

	agent := agent.New(
		llmClient,
	)

	app := cli.New(
		agent,
		executor,
	)

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
