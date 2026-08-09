package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aider/internal/executor"
	"aider/internal/llm"
	"aider/internal/models"
	"aider/internal/security"
)

const MaxIterations = 5

type Agent struct {
	llm              *llm.Client
	executor         *executor.Executor
	executionContext *executor.ExecutionContext
	policy           security.Policy
}

func New(
	llmClient *llm.Client,
	executor *executor.Executor,
	executionContext *executor.ExecutionContext,
	policy security.Policy,
) *Agent {
	return &Agent{
		llm:              llmClient,
		executor:         executor,
		executionContext: executionContext,
		policy:           policy,
	}
}

func (a *Agent) Explain(
	ctx context.Context,
	errorText string,
) (*models.Solution, error) {
	fmt.Println("[agent] Analyzing error...")
	fmt.Println()

	analysis, err := a.llm.Analyze(
		ctx,
		errorText,
	)
	if err != nil {
		return nil, err
	}

	solution := models.Solution{
		Problem:     analysis.Problem,
		Explanation: analysis.Explanation,
		Confidence:  analysis.Confidence,
		Risk:        analysis.Risk,
	}

	for _, action := range analysis.Actions {
		solution.Actions = append(
			solution.Actions,
			models.Action{
				Command: action.Command,
				Reason:  action.Reason,
			},
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
	agentContext *AgentContext,
	approve func(string, string) bool,
) error {

	for iteration := 1; iteration <= MaxIterations; iteration++ {

		fmt.Printf(
			"\n[agent] Iteration %d/%d\n\n",
			iteration,
			MaxIterations,
		)

		analysis, err := a.analyzeContext(
			ctx,
			agentContext,
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

			result := a.execute(
				ctx,
				command,
				args,
			)

			agentContext.AddStep(
				Step{
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
		`Current working directory:
%s


Command:
%s

Exit code:
%d

Stdout:
%s

Stderr:
%s`,
		a.executionContext.WorkingDirectory,
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

func (a *Agent) executeAction(
	ctx context.Context,
	action models.Action,
) executor.Result {

	if strings.HasPrefix(
		strings.TrimSpace(action.Command),
		"cd ",
	) {
		return a.changeDirectory(action.Command)
	}

	command, args, err := splitCommand(
		action.Command,
	)

	if err != nil {
		return executor.Result{
			Command:  action.Command,
			ExitCode: -1,
			Stderr:   err.Error(),
		}
	}

	return a.executor.Run(
		ctx,
		command,
		args...,
	)
}

func (a *Agent) changeDirectory(
	command string,
) executor.Result {

	parts := strings.Fields(command)

	if len(parts) != 2 {
		return executor.Result{
			Command:  command,
			ExitCode: -1,
			Stderr:   "invalid cd command",
		}
	}

	dir := parts[1]

	dir = expandHome(dir)

	if err := a.executionContext.SetWorkingDirectory(
		dir,
	); err != nil {
		return executor.Result{
			Command:  command,
			ExitCode: -1,
			Stderr:   err.Error(),
		}
	}

	return executor.Result{
		Command:  command,
		ExitCode: 0,
		Stdout: fmt.Sprintf(
			"Working directory changed to %s\n",
			a.executionContext.WorkingDirectory,
		),
	}
}

func expandHome(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(
				home,
				path[2:],
			)
		}
	}

	return path
}

func (a *Agent) analyzeContext(
	ctx context.Context,
	agentContext *AgentContext,
) (models.Analysis, error) {

	var builder strings.Builder

	fmt.Fprintf(
		&builder,
		`Original command:
%s

Current working directory:
%s

Execution history:
`,
		agentContext.OriginalCommand,
		agentContext.WorkingDirectory,
	)

	for i, step := range agentContext.History {

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

	return a.llm.Analyze(
		ctx,
		builder.String(),
	)
}
