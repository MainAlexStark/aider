package agent

import (
	"aider/internal/contexts"
	"aider/internal/executor"
	"aider/internal/llm"
	"aider/internal/models"
	"context"
	"fmt"
)

type Agent struct {
	llm          *llm.Client
	executor     *executor.Executor
	agentContext *contexts.AgentContext
	// policy           security.Policy
}

func New(
	llmClient *llm.Client,
	executor *executor.Executor,
	agentContext *contexts.AgentContext,
	// policy security.Policy,
) *Agent {
	return &Agent{
		llm:          llmClient,
		executor:     executor,
		agentContext: agentContext,
		// policy:           policy,
	}
}

func (a *Agent) Explain(
	ctx context.Context,
	errorText string,
) (*models.Solution, error) {
	fmt.Println("[agent] Analyzing error...")

	// Отправляем запрос на llm
	solution, err := a.llm.Analize(ctx, errorText, a.agentContext)
	if err != nil {
		return nil, err
	}

	return &solution, nil
}
