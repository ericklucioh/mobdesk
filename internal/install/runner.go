package install

import (
	"bytes"
	"context"

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
	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdout, err := command.Output()
	if err == nil {
		return CommandResult{Stdout: stdout, Stderr: stderr.Bytes()}
	}
	return CommandResult{Err: err, Stdout: stdout, Stderr: stderr.Bytes()}
}
