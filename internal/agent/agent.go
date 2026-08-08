package agent

import (
	"aider/internal/executor"
	"aider/internal/llm"
	"fmt"
)

type Agent struct {
	llm *llm.Client
}

func New(llmClient *llm.Client) *Agent {
	return &Agent{
		llm: llmClient,
	}
}

func (a *Agent) Explain(errorText string) error {
	fmt.Println("[agent] Analyzing error...")
	fmt.Println()

	response, err := a.llm.Analyze(errorText)
	if err != nil {
		return err
	}

	fmt.Println(response)

	return nil
}

func (a *Agent) Analyze(result executor.Result) error {
	fmt.Println("[agent] Analyzing error...")
	fmt.Println()

	errorText := fmt.Sprintf(
		`Command:
%s

Exit code:
%d

Stdout:
%s

Stderr:
%s`,
		result.Command,
		result.ExitCode,
		result.Stdout,
		result.Stderr,
	)

	response, err := a.llm.Analyze(errorText)
	if err != nil {
		return err
	}

	fmt.Println(response)

	return nil
}
