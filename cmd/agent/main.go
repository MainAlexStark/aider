package main

import (
	"aider/internal/agent"
	"aider/internal/cli"
	"aider/internal/config"
	"aider/internal/contexts"
	"aider/internal/executor"
	"aider/internal/llm"
	"aider/internal/security"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Прежде всего должен открыватся -h and --help
	if len(os.Args) >= 2 &&
		(os.Args[1] == "-h" || os.Args[1] == "--help") {

		cli.PrintHelp()
		return
	}

	// Загружаем конфиг
	_ = godotenv.Load()
	cnf, err := config.Load()
	if err != nil {
		cli.PrintConfigHelp()
		panic(err)
	}
	_ = cnf

	// Инициализируем контекст агента
	agentContext, err := contexts.NewAgentContext()
	if err != nil {
		fmt.Fprintln(
			os.Stderr,
			"error:",
			err,
		)
		os.Exit(1)
	}

	// Инициализируем executor
	exec := executor.New(
		agentContext,
	)

	// Инициализируем клиент OpenRouter
	llmClient := llm.New(
		agentContext,
		cnf.OpenRouterAPIKey,
		cnf.Model,
		cnf.Language,
	)

	// Инициализируем правила
	policy := security.Policy{
		MaxCommandLength:    4096,
		MaxOutputSize:       32 * 1024,
		AllowShell:          false,
		AllowSudo:           false,
		AllowNetwork:        true,
		AllowPackageInstall: true,
	}

	// Иницализируем агента
	agent := agent.New(
		llmClient,
		exec,
		agentContext,
		cnf.MaxIterations,
		policy,
	)

	// Запускаем cli приложенения
	app := cli.New(
		agent,
		exec,
		agentContext,
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
