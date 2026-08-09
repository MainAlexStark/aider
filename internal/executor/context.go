package executor

import (
	"fmt"
	"os"
	"path/filepath"
)

type ExecutionContext struct {
	WorkingDirectory string
}

func NewExecutionContext() (*ExecutionContext, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf(
			"get working directory: %w",
			err,
		)
	}

	return &ExecutionContext{
		WorkingDirectory: dir,
	}, nil
}

func (c *ExecutionContext) SetWorkingDirectory(
	dir string,
) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf(
			"resolve working directory: %w",
			err,
		)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf(
			"working directory does not exist: %w",
			err,
		)
	}

	if !info.IsDir() {
		return fmt.Errorf(
			"working directory is not a directory: %s",
			abs,
		)
	}

	c.WorkingDirectory = abs

	return nil
}
