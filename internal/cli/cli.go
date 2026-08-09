package cli

import (
	"aider/internal/agent"
	"aider/internal/executor"
	"context"
	"fmt"
	"os"
	"strings"
)

type CLI struct {
	agent    agent.Agent
	executor *executor.Executor
}

func New(
	agent agent.Agent,
	executor *executor.Executor,
) *CLI {
	return &CLI{
		agent:    agent,
		executor: executor,
	}
}

func (c *CLI) Run(args []string) error {
	if len(args) < 2 {
		c.printUsage()
		return nil
	}

	mode := args[1]

	switch mode {
	case "explain":
		return c.explain(args[2:])

	case "run":
		return c.run(args[2:])

	default:
		return fmt.Errorf("unknown mode: %s", mode)
	}
}

func (c *CLI) explain(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("error message is required")
	}

	errorText := strings.Join(args, " ")

	return c.agent.Explain(errorText)
}

func (c *CLI) run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("command is required")
	}

	command := args[0]
	commandArgs := args[1:]

	fmt.Printf("$ %s %s\n\n", command, strings.Join(commandArgs, " "))

	result := c.executor.Run(
		context.Background(),
		command,
		commandArgs...,
	)

	fmt.Print(result.Stdout)
	fmt.Print(result.Stderr)

	if result.ExitCode == 0 {
		fmt.Println("\n✓ Command completed successfully")
		return nil
	}

	fmt.Printf(
		"\nCommand failed with exit code %d\n",
		result.ExitCode,
	)

	return c.agent.Analyze(result)
}

func (c *CLI) printUsage() {
	fmt.Println(`
	Terminal Agent

	Usage:

	agent explain "<error>"
	agent run <command> [args...]

	Examples:

	agent explain "connection refused"

	agent run docker compose build

	agent run go test ./...
	`)
}

func Execute() {
	// TODO
	_ = os.Args
}
