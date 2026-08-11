package models

import (
	"fmt"
)

// ExecutorResult
type Result struct {
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
}

// Agent explain result
type Risk string

const (
	RiskLow      Risk = "low"
	RiskMedium   Risk = "medium"
	RiskHigh     Risk = "high"
	RiskCritical Risk = "critical"
)

type Action struct {
	Command string `json:"command"`
	Reason  string `json:"reason"`
	Risk    Risk   `json:"risk"`
}

type Solution struct {
	Problem     string   `json:"problem"`
	Explanation string   `json:"explanation"`
	Confidence  float64  `json:"confidence"`
	Risk        Risk     `json:"risk"`
	Actions     []Action `json:"actions"`
}

func (solution *Solution) PrintSolution() {
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
