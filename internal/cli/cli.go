package cli

import (
	"aider/internal/config"
	"fmt"

	"github.com/joho/godotenv"
)

type CLI struct {
	// agent *agent.Agent
	// executor         *executor.Executor
	// executionContext *executor.ExecutionContext
	// policy           security.Policy
}

func New(
// agent *agent.Agent,
// executor *executor.Executor,
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
