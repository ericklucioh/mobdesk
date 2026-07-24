package status

import (
	"context"
	"os/exec"

	"github.com/ericklucioh/mobdesk/internal/executil"
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
	command, err := executil.CommandContext(ctx, name, args...)
	if err != nil {
		return CommandResult{Err: err}
	}
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
