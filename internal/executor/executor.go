package executor

import (
	"aider/internal/models"
	"context"
	"io"
	"os/exec"
)

type Executor struct {
	ctx *ExecutionContext
}

func New(ctx *ExecutionContext) *Executor {
	return &Executor{
		ctx: ctx,
	}
}

func (e *Executor) Run(
	ctx context.Context,
	command string,
	args ...string,
) models.Result {

	cmd := exec.CommandContext(
		ctx,
		command,
		args...,
	)

	cmd.Dir = e.ctx.WorkingDirectory

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return models.Result{
			Command:  command,
			ExitCode: -1,
			Stderr:   err.Error(),
		}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return models.Result{
			Command:  command,
			ExitCode: -1,
			Stderr:   err.Error(),
		}
	}

	if err := cmd.Start(); err != nil {
		return models.Result{
			Command:  command,
			ExitCode: -1,
			Stderr:   err.Error(),
		}
	}

	stdoutBytes, _ := io.ReadAll(stdout)
	stderrBytes, _ := io.ReadAll(stderr)

	err = cmd.Wait()

	exitCode := 0

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return models.Result{
		Command:  command,
		ExitCode: exitCode,
		Stdout:   string(stdoutBytes),
		Stderr:   string(stderrBytes),
	}
}
