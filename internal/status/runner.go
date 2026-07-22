package status

import (
	"context"
	"os/exec"
)

type CommandResult struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) CommandResult
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) CommandResult {
	command := exec.CommandContext(ctx, name, args...)
	stdout, err := command.Output()
	if err == nil {
		return CommandResult{Stdout: stdout}
	}
	result := CommandResult{Err: err, Stdout: stdout}
	if exitError, ok := err.(*exec.ExitError); ok {
		result.Stderr = exitError.Stderr
	}
	return result
}
