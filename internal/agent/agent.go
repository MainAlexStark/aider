package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"aider/internal/executor"
	"aider/internal/llm"
	"aider/internal/security"
)

type Agent struct {
	llm *llm.Client
}

func New(llmClient *llm.Client) Agent {
	return Agent{
		llm: llmClient,
	}
}

func (a Agent) Explain(
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
	ctx context.Context,
	result executor.Result,
) error {

	fmt.Println("[agent] Analyzing error...")
	fmt.Println()

	stdout := security.Redact(
		security.LimitOutput(
			result.Stdout,
			32*1024,
		),
	)

	stderr := security.Redact(
		security.LimitOutput(
			result.Stderr,
			32*1024,
		),
	)

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
		stdout,
		stderr,
	)

	response, err := a.llm.Analyze(
		ctx,
		errorText,
	)

	if err != nil {
		return err
	}

	fmt.Println(response)

	return nil
}
