package agent

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

func (c *AgentContext) AddStep(
	step Step,
) {
	c.History = append(
		c.History,
		step,
	)
}
