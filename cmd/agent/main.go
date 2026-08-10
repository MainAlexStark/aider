package main

import (
	"aider/internal/cli"
	"aider/internal/config"
	"aider/internal/executor"
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

	// Инициализируем контекст для executor
	executionContext, err := executor.NewExecutionContext()
	if err != nil {
		fmt.Fprintln(
			os.Stderr,
			"failed to create execution context:",
			err,
		)

		os.Exit(1)
	}
	// Инициализируем executor
	exec := executor.New(
		executionContext,
	)

	// Запускаем cli приложенения
	app := cli.New(exec)
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(
			os.Stderr,
			"error:",
			err,
		)

		os.Exit(1)
	}
}
