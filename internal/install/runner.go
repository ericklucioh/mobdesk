package install

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/creack/pty/v2"
	"github.com/ericklucioh/mobdesk/internal/executil"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
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
	command.WaitDelay = 500 * time.Millisecond
	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdout, err := command.Output()
	return CommandResult{Stdout: stdout, Stderr: stderr.Bytes(), Err: err}
}

// InteractiveRunner gives native package managers terminal ownership while
// retaining their output for the installation log.
type InteractiveRunner struct{}

func (InteractiveRunner) Run(ctx context.Context, name string, args ...string) CommandResult {
	command, err := executil.CommandContext(ctx, name, args...)
	if err != nil {
		return CommandResult{Err: err}
	}
	if _, err := fmt.Fprintf(os.Stdout, "\n$ %s %s\n", name, strings.Join(args, " ")); err != nil {
		return CommandResult{Err: err}
	}
	ptmx, err := pty.Start(command)
	if err != nil {
		return CommandResult{Err: err}
	}
	defer func() { _ = ptmx.Close() }()
	_ = pty.InheritSize(os.Stdin, ptmx)

	fd := int(os.Stdin.Fd())
	var terminalState *term.State
	if term.IsTerminal(fd) {
		terminalState, err = term.MakeRaw(fd)
		if err != nil {
			_ = command.Process.Kill()
			_ = command.Wait()
			return CommandResult{Err: err}
		}
		defer func() { _ = term.Restore(fd, terminalState) }()
	}

	var output bytes.Buffer
	outputDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.MultiWriter(os.Stdout, &output), ptmx)
		close(outputDone)
	}()
	inputContext, cancelInput := context.WithCancel(ctx)
	inputDone := make(chan struct{})
	go func() {
		defer close(inputDone)
		copyTerminalInput(inputContext, ptmx, os.Stdin)
	}()

	waitErr := command.Wait()
	cancelInput()
	_ = ptmx.Close()
	<-outputDone
	<-inputDone
	if waitErr == nil && ctx.Err() != nil {
		waitErr = ctx.Err()
	}
	return CommandResult{Stdout: output.Bytes(), Err: waitErr}
}

func copyTerminalInput(ctx context.Context, destination io.Writer, source *os.File) {
	buffer := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		poll := []unix.PollFd{{Fd: int32(source.Fd()), Events: unix.POLLIN}}
		_, err := unix.Poll(poll, 100)
		if err == unix.EINTR {
			continue
		}
		if err != nil || poll[0].Revents&(unix.POLLERR|unix.POLLHUP|unix.POLLNVAL) != 0 {
			return
		}
		if poll[0].Revents&unix.POLLIN == 0 {
			continue
		}
		count, err := source.Read(buffer)
		if count > 0 {
			if _, writeErr := destination.Write(buffer[:count]); writeErr != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func runnerFor(options Options) CommandRunner {
	if options.Runner != nil {
		return options.Runner
	}
	if options.Interactive {
		return InteractiveRunner{}
	}
	return ExecRunner{}
}
