package cli

import (
	"aider/internal/agent"
	"aider/internal/config"
	"aider/internal/executor"
	"aider/internal/models"
	"aider/internal/security"
	"context"
	"fmt"
	"strings"

	"github.com/joho/godotenv"
)

type CLI struct {
	agent            *agent.Agent
	executor         *executor.Executor
	executionContext *executor.ExecutionContext
	policy           security.Policy
}

func New(
	agent *agent.Agent,
	executor *executor.Executor,
	executionContext *executor.ExecutionContext,
	policy security.Policy,
) *CLI {
	return &CLI{
		agent:            agent,
		executor:         executor,
		executionContext: executionContext,
		policy:           policy,
	}
}

func (c *CLI) Run(args []string) error {
	if len(args) < 2 {
		PrintHelp()
		return nil
	}

	switch args[1] {

	case "explain":
		if len(args) >= 3 &&
			(args[2] == "-h" || args[2] == "--help") {
			c.printExplainHelp()
			return nil
		}

		return c.explain(args[2:])

	case "run":
		if len(args) >= 3 &&
			(args[2] == "-h" || args[2] == "--help") {
			c.printRunHelp()
			return nil
		}

		return c.run(args[2:])

	case "config":
		return c.config(args[2:])

	default:
		return fmt.Errorf("unknown mode: %s", args[1])
	}
}

func (c *CLI) explain(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("error message is required")
	}

	errorText := strings.Join(args, " ")

	solution, err := c.agent.Explain(context.Background(), errorText)
	if err != nil {
		return err
	}

	printSolution(solution)

	return nil
}

func (c *CLI) run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("command is required")
	}

	command := strings.Join(args, " ")

	// Проверяем исходную команду через security policy.
	decision, reason := security.Validate(
		command,
		c.policy,
	)

	switch decision {

	case security.DecisionBlock:
		fmt.Println()
		fmt.Println("[SECURITY BLOCKED]")
		fmt.Println(reason)

		return nil

	case security.DecisionApproval:
		fmt.Println()
		fmt.Println("[SECURITY WARNING]")
		fmt.Println(reason)

		if !confirm() {
			fmt.Println("Execution cancelled.")
			return nil
		}
	}

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
			"\n✓ Command completed successfully",
		)

		return nil
	}

	fmt.Printf(
		"\nCommand failed with exit code %d\n",
		result.ExitCode,
	)

	// Создаём состояние агента.
	agentContext := &agent.AgentContext{
		OriginalCommand:  command,
		WorkingDirectory: c.executionContext.WorkingDirectory,
	}

	// Добавляем исходную команду в историю.
	agentContext.AddStep(
		agent.Step{
			Command:  result.Command,
			ExitCode: result.ExitCode,
			Stdout:   result.Stdout,
			Stderr:   result.Stderr,
		},
	)

	// Передаём управление агенту.
	return c.agent.Run(
		ctx,
		agentContext,
		approve,
	)

}

func confirm() bool {
	fmt.Print("\nExecute command? [y/N]: ")

	var answer string

	_, err := fmt.Scanln(&answer)
	if err != nil {
		return false
	}

	answer = strings.ToLower(
		strings.TrimSpace(answer),
	)

	return answer == "y" ||
		answer == "yes"
}

func (c *CLI) printRunHelp() {
	fmt.Println(`
	Usage:
	aider run <command> [arguments...]

	Description:
	Execute a terminal command.
	If the command fails, Aider analyzes the error
	and suggests a solution.

	Security:
	Commands are checked by the security layer before execution.
	Potentially dangerous commands require confirmation.
	Some commands are blocked completely.

	Examples:
	aider run echo hello
	aider run go test ./...
	aider run docker compose build
	aider run git status

	Options:
	-h, --help    Show this help message
	`)
}

func (c *CLI) printExplainHelp() {
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

	Options:
	-h, --help    Show this help message

	Examples:
	aider run echo hello
	aider run go test ./...
	aider run docker compose build

	aider explain "connection refused"
	`)
}

func printSolution(solution *models.Solution) {
	fmt.Println()
	fmt.Println("══════════════════════════════════════")
	fmt.Println("              ANALYSIS")
	fmt.Println("══════════════════════════════════════")

	fmt.Printf("\nProblem:\n%s\n", solution.Problem)

	fmt.Printf(
		"\nExplanation:\n%s\n",
		solution.Explanation,
	)

	fmt.Printf(
		"\nConfidence: %.0f%%\n",
		solution.Confidence*100,
	)

	fmt.Printf(
		"Risk: %s\n",
		solution.Risk,
	)

	if len(solution.Actions) == 0 {
		fmt.Println("\nNo actions suggested.")
		return
	}

	fmt.Println("\nSuggested actions:")

	for i, action := range solution.Actions {
		fmt.Printf(
			"\n[%d] %s\n",
			i+1,
			action.Command,
		)

		fmt.Printf(
			"    %s\n",
			action.Reason,
		)
	}
}

func (c *CLI) config(args []string) error {
	if len(args) == 0 {
		c.printConfigHelp()
		return nil
	}

	switch args[0] {

	case "-h", "--help":
		c.printConfigHelp()
		return nil

	case "path":
		return c.configPath()

	case "show":
		return c.configShow()

	case "set":
		return c.configSet(args[1:])

	default:
		return fmt.Errorf(
			"unknown config command: %s",
			args[0],
		)
	}
}

func (c *CLI) configPath() error {
	path, err := config.Path()
	if err != nil {
		return err
	}

	fmt.Println(path)

	return nil
}

func (c *CLI) configShow() error {
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

func (c *CLI) printConfigHelp() {
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

func approve(
	command string,
	reason string,
) bool {

	fmt.Println()
	fmt.Println("──────────────────────────────────────")
	fmt.Println("AGENT ACTION")
	fmt.Println("──────────────────────────────────────")

	fmt.Printf(
		"Command: %s\n",
		command,
	)

	fmt.Printf(
		"Reason: %s\n",
		reason,
	)

	fmt.Print(
		"\nExecute this command? [y/N]: ",
	)

	var answer string

	if _, err := fmt.Scanln(&answer); err != nil {
		return false
	}

	answer = strings.ToLower(
		strings.TrimSpace(answer),
	)

	return answer == "y" ||
		answer == "yes"
}
