package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"aider/internal/executor"
	"aider/internal/llm"
	"aider/internal/models"
	"aider/internal/security"
)

const MaxIterations = 5

type Agent struct {
	llm      *llm.Client
	policy   security.Policy
	executor *executor.Executor
}

func New(
	llmClient *llm.Client,
	executor *executor.Executor,
	policy security.Policy,
) *Agent {
	return &Agent{
		llm:      llmClient,
		executor: executor,
		policy:   policy,
	}
}

func (a Agent) Explain(
	ctx context.Context,
	errorText string,
) (*models.Solution, error) {
	fmt.Println("[agent] Analyzing error...")
	fmt.Println()

	response, err := a.llm.Analyze(
		context.Background(),
		errorText,
	)
	if err != nil {
		return nil, err
	}

	var solution models.Solution

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

func (a *Agent) Run(
	ctx context.Context,
	result executor.Result,
	approve func(string, string) bool,
) error {

	for iteration := 1; iteration <= MaxIterations; iteration++ {

		fmt.Printf(
			"\n[agent] Iteration %d/%d\n\n",
			iteration,
			MaxIterations,
		)

		analysis, err := a.analyzeResult(
			ctx,
			result,
		)

		if err != nil {
			return err
		}

		a.printAnalysis(analysis)

		if len(analysis.Actions) == 0 {
			fmt.Println(
				"[agent] No further actions suggested.",
			)

			return nil
		}

		executed := false

		for _, action := range analysis.Actions {

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

			result = a.execute(
				ctx,
				command,
				args,
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
		MaxIterations,
	)

	return nil
}

func (a *Agent) execute(
	ctx context.Context,
	command string,
	args []string,
) executor.Result {

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

func (a *Agent) analyzeResult(
	ctx context.Context,
	result executor.Result,
) (models.Analysis, error) {

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
		security.Redact(
			security.LimitOutput(
				result.Stdout,
				a.policy.MaxOutputSize,
			),
		),
		security.Redact(
			security.LimitOutput(
				result.Stderr,
				a.policy.MaxOutputSize,
			),
		),
	)

	return a.llm.Analyze(
		ctx,
		errorText,
	)
}

func (a *Agent) printAnalysis(
	analysis models.Analysis,
) {

	fmt.Println("══════════════════════════════════════")
	fmt.Println("ANALYSIS")
	fmt.Println("══════════════════════════════════════")

	fmt.Printf(
		"Problem:\n%s\n\n",
		analysis.Problem,
	)

	fmt.Printf(
		"Explanation:\n%s\n\n",
		analysis.Explanation,
	)

	fmt.Printf(
		"Confidence: %.0f%%\n",
		analysis.Confidence*100,
	)

	fmt.Printf(
		"Risk: %s\n",
		analysis.Risk,
	)

	if len(analysis.Actions) > 0 {

		fmt.Println()
		fmt.Println("Suggested actions:")

		for i, action := range analysis.Actions {

			fmt.Printf(
				"\n%d. %s\n",
				i+1,
				action.Command,
			)

			fmt.Printf(
				"   %s\n",
				action.Reason,
			)
		}
	}
}
