package install

import (
	"bytes"
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
	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdout, err := command.Output()
	if err == nil {
		return CommandResult{Stdout: stdout, Stderr: stderr.Bytes()}
	}
	return CommandResult{Err: err, Stdout: stdout, Stderr: stderr.Bytes()}
}
