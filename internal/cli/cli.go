package cli

import (
	"aider/internal/config"
	"aider/internal/executor"
	"aider/internal/models"
	"context"
	"fmt"
	"strings"

	"github.com/joho/godotenv"
)

type CLI struct {
	// agent *agent.Agent
	executor *executor.Executor
	// executionContext *executor.ExecutionContext
	// policy           security.Policy
}

func New(
	// agent *agent.Agent,
	executor *executor.Executor,
	// executionContext *executor.ExecutionContext,
	// policy security.Policy,
) *CLI {
	return &CLI{
		// agent: agent,
		// executor:         executor,
		// executionContext: executionContext,
		// policy:           policy,
	}
}

func (c *CLI) Run(args []string) error {
	// Нет аргументов - отправляем помощь
	if len(args) < 2 {
		PrintHelp()
		return nil
	}

	switch args[1] {

	case "explain":
		if len(args) >= 3 &&
			(args[2] == "-h" || args[2] == "--help") {
			printExplainHelp()
			return nil
		}

		return c.explain(args[2:])

	case "explainplus":
		if len(args) >= 3 &&
			(args[2] == "-h" || args[2] == "--help") {
			printExplainPlusHelp()
			return nil
		}
		return c.explainPlus(args[2:])

	case "config":
		return c.config(args[2:])

	default:
		return fmt.Errorf("unknown mode: %s, use -h or --help for more information", args[1])
	}
}

func PrintHelp() {
	fmt.Println(`
	Aider - terminal AI agent

	Usage:
	aider <command> [arguments]

	Commands:
	run       Execute a command and analyze errors
	explain   Explain an error using AI
	config    Manage Aider configuration

	aider config show
	aider config set language ru
	aider config set api-key
	aider config set model nvidia/nemotron-3-ultra-550b-a55b:free
	aider config set max_iterations 5

	Options:
	-h, --help    Show this help message

	Examples:
	aider run echo hello
	aider run go test ./...
	aider run docker compose build

	aider explain "connection refused"
	`)
}

func PrintConfigHelp() {
	fmt.Println(`
	Usage:
	aider config <command>

	Commands:
	show                 Show current configuration
	path                 Show configuration file path
	set language <lang>  Set agent language
	set model <model>    Set agent model
	set api-key          Set OpenRouter API key
	set max_iterations	 Set maximum iterations count

	Examples:
	aider config show
	aider config path
	aider config set language ru
	aider config set language en
	aider config set api-key

	Options:
	-h, --help           Show this help message
	`)
}

func printExplainHelp() {
	fmt.Println(`
	Usage:
	aider explain "<error>"

	Description:
	Analyze an error using the configured AI model
	and provide an explanation and possible solution.

	Examples:
	aider explain "connection refused"

	aider explain "docker compose build fails with TLS handshake timeout"

	aider explain "go test fails with undefined: foo"

	Options:
	-h, --help    Show this help message
	`)
}

func printExplainPlusHelp() {
	fmt.Println(`
	Usage:
	aider explainplus <command>

	Description:
	Execute a command, analyze its output and errors using the configured AI model,
	and provide an explanation and possible solution.

	Examples:
	aider explainplus docker compose up

	Options:
	-h, --help    Show this help message
	`)
}

// Run command
func (c *CLI) run_command(args []string) (models.Result, error) {
	if len(args) == 0 {
		return models.Result{}, fmt.Errorf("command is required")
	}

	command := strings.Join(args, " ")
	// TODO добавить проверку на запрещенные команды
	_ = command

	commandName := args[0]
	commandArgs := args[1:]

	fmt.Printf(
		"$ %s %s\n\n",
		commandName,
		strings.Join(commandArgs, " "),
	)

	// Выполняем исходную команду.
	ctx := context.Background()

	result := c.executor.Run(
		ctx,
		commandName,
		commandArgs...,
	)

	fmt.Print(result.Stdout)
	fmt.Print(result.Stderr)

	// Команда выполнилась успешно.
	if result.ExitCode == 0 {
		fmt.Println(
			"\n✅Command completed successfully",
		)

		return models.Result{}, nil
	}

	return result, nil
}

// Explain mode commands
func (c *CLI) explain(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("error message is required")
	}

	errorText := strings.Join(args, " ")
	_ = errorText

	return nil
}

func (c *CLI) explainPlus(args []string) error {
	// Выполняем команду и получаем результат
	result, err := c.run_command(args)
	if err != nil {
		return err
	}

	// Если команда выполнилась успешно - объяснять нечего.
	if result.ExitCode == 0 {
		return nil
	}

	return c.explain([]string{result.Stderr})
}

// config handles the "config" command and its subcommands.
func (c *CLI) config(args []string) error {
	// Если нет аргументов - возвращаем помощь
	if len(args) == 0 {
		PrintConfigHelp()
		return nil
	}

	switch args[0] {

	case "-h", "--help":
		PrintConfigHelp()
		return nil

	case "path":
		return c.printConfigPath()

	case "show":
		return c.printConfigData()

	case "set":
		return c.configSet(args[1:])

	default:
		return fmt.Errorf(
			"unknown config command: %s",
			args[0],
		)
	}
}

func (c *CLI) printConfigPath() error {
	path, err := config.Path()
	if err != nil {
		return err
	}

	fmt.Println(path)

	return nil
}

func (c *CLI) printConfigData() error {
	path, err := config.Path()
	if err != nil {
		return err
	}

	values, err := godotenv.Read(path)
	if err != nil {
		return err
	}

	fmt.Println("Configuration:")
	fmt.Println()

	for key, value := range values {
		if key == "OPENROUTER_API_KEY" {
			value = maskSecret(value)
		}

		fmt.Printf(
			"%s=%s\n",
			key,
			value,
		)
	}

	return nil
}

func maskSecret(value string) string {
	if len(value) <= 8 {
		return "********"
	}

	return value[:4] +
		"********" +
		value[len(value)-4:]
}

func (c *CLI) configSet(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf(
			"configuration key is required",
		)
	}

	switch args[0] {

	case "language":
		if len(args) < 2 {
			return fmt.Errorf(
				"language is required",
			)
		}

		return config.Set(
			"AGENT_LANGUAGE",
			args[1],
		)

	case "model":
		if len(args) < 2 {
			return fmt.Errorf(
				"model is required",
			)
		}

		return config.Set(
			"AGENT_MODEL",
			args[1],
		)

	case "max_iterations":
		if len(args) < 2 {
			return fmt.Errorf(
				"max_iterations is required",
			)
		}

		return config.Set(
			"MAX_ITERATIONS",
			args[1],
		)

	case "api-key":
		return c.setAPIKey()

	default:
		return fmt.Errorf(
			"unknown configuration key: %s",
			args[0],
		)
	}
}

func (c *CLI) setAPIKey() error {
	fmt.Print("OpenRouter API key: ")

	var apiKey string

	_, err := fmt.Scanln(&apiKey)
	if err != nil {
		return err
	}

	if apiKey == "" {
		return fmt.Errorf(
			"API key cannot be empty",
		)
	}

	if err := config.Set(
		"OPENROUTER_API_KEY",
		apiKey,
	); err != nil {
		return err
	}

	fmt.Println(
		"OpenRouter API key saved.",
	)

	return nil
}
