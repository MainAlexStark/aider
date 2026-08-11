package agent

import (
	"aider/internal/contexts"
	"aider/internal/executor"
	"aider/internal/llm"
	"aider/internal/models"
	"aider/internal/security"
	"context"
	"fmt"
	"strings"
)

type Agent struct {
	llm            *llm.Client
	executor       *executor.Executor
	agentContext   *contexts.AgentContext
	max_iterations int
	policy         security.Policy
}

func New(
	llmClient *llm.Client,
	executor *executor.Executor,
	agentContext *contexts.AgentContext,
	max_iterations int,
	policy security.Policy,
) *Agent {
	return &Agent{
		llm:            llmClient,
		executor:       executor,
		agentContext:   agentContext,
		max_iterations: max_iterations,
		policy:         policy,
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

func (a *Agent) Run(
	ctx *context.Context,
	approve func(string, string) bool,
) error {
	for iteration := 1; iteration <= a.max_iterations; iteration++ {

		fmt.Printf(
			"\n[agent] Iteration %d/%d\n\n",
			iteration,
			a.max_iterations,
		)

		// Подаем в LLM всю историю целиком
		solution, err := a.llm.Analize(
			*ctx,
			a.getAnalyzeContextStr(),
			a.agentContext,
		)
		if err != nil {
			return err
		}

		solution.PrintSolution()

		// Если предлагаемых решений нет, то завершаемся
		if len(solution.Actions) == 0 {
			fmt.Println(
				"[agent] No further actions suggested.",
			)

			return nil
		}

		executed := false

		for _, action := range solution.Actions {

			// Проверки безопастности
			decision, reason := security.Validate(
				action.Command,
				a.policy,
			)

			switch decision {

			case security.DecisionBlock:

				fmt.Printf(
					"\n[SECURITY BLOCKED]\n%s\n",
					reason,
				)

				continue

			case security.DecisionApproval:

				if !approve(
					action.Command,
					action.Reason,
				) {
					fmt.Println(
						"[agent] Action rejected.",
					)

					continue
				}

			case security.DecisionAllow:
				// Nothing to do.
			}

			fmt.Printf(
				"\n[agent] Executing: %s\n",
				action.Command,
			)

			command, args, err := splitCommand(
				action.Command,
			)

			if err != nil {
				fmt.Printf(
					"[agent] Invalid command: %v\n",
					err,
				)

				continue
			}

			result := a.execute(
				*ctx,
				command,
				args,
			)

			a.agentContext.AddStep(
				contexts.Step{
					Command:  result.Command,
					ExitCode: result.ExitCode,
					Stdout:   result.Stdout,
					Stderr:   result.Stderr,
				},
			)

			fmt.Print(result.Stdout)
			fmt.Print(result.Stderr)

			if result.ExitCode == 0 {
				fmt.Println(
					"[agent] Action completed successfully.",
				)
			} else {
				fmt.Printf(
					"[agent] Action failed with exit code %d.\n",
					result.ExitCode,
				)
			}

			executed = true

			// После выполнения одной команды
			// возвращаемся к LLM.
			break

		}

		if !executed {
			fmt.Println(
				"[agent] No actions were executed.",
			)

			return nil
		}
	}

	fmt.Printf(
		"\n[agent] Maximum iterations (%d) reached.\n",
		a.max_iterations,
	)

	return nil
}

func (a *Agent) execute(
	ctx context.Context,
	command string,
	args []string,
) models.Result {

	return a.executor.Run(
		ctx,
		command,
		args...,
	)
}

func splitCommand(
	command string,
) (string, []string, error) {

	parts := strings.Fields(command)

	if len(parts) == 0 {
		return "", nil, fmt.Errorf(
			"empty command",
		)
	}

	return parts[0], parts[1:], nil
}

func (a *Agent) getAnalyzeContextStr() string {
	var builder strings.Builder

	fmt.Fprintf(
		&builder,
		`Original command:
%s

Current working directory:
%s

Execution history:
`,
		a.agentContext.OriginalCommand,
		a.agentContext.WorkingDirectory,
	)

	for i, step := range a.agentContext.History {

		fmt.Fprintf(
			&builder,
			`
--- Step %d ---
Command:
%s

Exit code:
%d

Stdout:
%s

Stderr:
%s
`,
			i+1,
			step.Command,
			step.ExitCode,
			security.Redact(
				security.LimitOutput(
					step.Stdout,
					a.policy.MaxOutputSize,
				),
			),
			security.Redact(
				security.LimitOutput(
					step.Stderr,
					a.policy.MaxOutputSize,
				),
			),
		)
	}
	return builder.String()
}
