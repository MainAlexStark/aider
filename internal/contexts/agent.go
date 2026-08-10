package contexts

import (
	"fmt"
	"os"
	"strings"
)

type AgentContext struct {
	OriginalCommand  string
	WorkingDirectory string
	History          []Step
}

type Step struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
}

func NewAgentContext() (*AgentContext, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf(
			"get working directory: %w",
			err,
		)
	}

	return &AgentContext{
		WorkingDirectory: dir,
	}, nil
}

func (c *AgentContext) AddStep(
	step Step,
) {
	c.History = append(
		c.History,
		step,
	)
}

func (c *AgentContext) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Original command:\n%s\n\n", c.OriginalCommand)
	fmt.Fprintf(&b, "Working directory:\n%s\n", c.WorkingDirectory)

	if len(c.History) == 0 {
		return b.String()
	}

	b.WriteString("\nExecution history:\n")

	for i, step := range c.History {
		fmt.Fprintf(&b, "\n--- Step %d ---\n", i+1)

		fmt.Fprintf(&b, "Command:\n%s\n", step.Command)
		fmt.Fprintf(&b, "Exit code: %d\n", step.ExitCode)

		b.WriteString("Stdout:\n")
		if step.Stdout != "" {
			b.WriteString(step.Stdout)
			b.WriteString("\n")
		}

		b.WriteString("Stderr:\n")
		if step.Stderr != "" {
			b.WriteString(step.Stderr)
			b.WriteString("\n")
		}
	}

	return b.String()
}
