package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"aider/internal/executor"
	"aider/internal/llm"
)

type Agent struct {
	llm *llm.Client
}

func New(llmClient *llm.Client) Agent {
	return Agent{
		llm: llmClient,
	}
}

func (a *Agent) Explain(
	ctx context.Context,
	errorText string,
) (*Solution, error) {
	fmt.Println("[agent] Analyzing error...")
	fmt.Println()

	response, err := a.llm.Analyze(
		context.Background(),
		errorText,
	)
	if err != nil {
		return nil, err
	}

	var solution Solution

	if err := json.Unmarshal(
		[]byte(response),
		&solution,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to parse LLM response: %w\nresponse: %s",
			err,
			response,
		)
	}

	return &solution, nil
}

func (a *Agent) Analyze(
	result executor.Result,
) (*Solution, error) {
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

	return a.Explain(context.Background(), errorText)
}
