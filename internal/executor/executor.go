package executor

import (
	"bytes"
	"context"
	"os/exec"
)

type Result struct {
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
}

type Executor struct{}

func New() *Executor {
	return &Executor{}
}

func (e *Executor) Run(
	ctx context.Context,
	command string,
	args ...string,
) Result {
	cmd := exec.CommandContext(ctx, command, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := Result{
		Command:  command,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}
