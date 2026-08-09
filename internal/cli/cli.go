package cli

import (
	"aider/internal/agent"
	"aider/internal/executor"
	"aider/internal/security"
	"context"
	"fmt"
	"os"
	"strings"
)

type CLI struct {
	agent    *agent.Agent
	executor *executor.Executor
	policy   security.Policy
}

func New(
	agent *agent.Agent,
	executor *executor.Executor,
	policy security.Policy,
) *CLI {
	return &CLI{
		agent:    agent,
		executor: executor,
		policy:   policy,
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

	result := c.executor.Run(
		context.Background(),
		commandName,
		commandArgs...,
	)

	fmt.Print(result.Stdout)
	fmt.Print(result.Stderr)

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

	return c.agent.Analyze(
		context.Background(),
		result,
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

func printSolution(solution *agent.Solution) {
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
