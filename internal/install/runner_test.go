package install

import (
	"context"
	"errors"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

type recordingRunner struct {
	name string
	args []string
}

func (r *recordingRunner) Run(_ context.Context, name string, args ...string) CommandResult {
	r.name = name
	r.args = append([]string(nil), args...)
	return CommandResult{}
}

func TestNonInteractiveRunnerAddsUbuntuEnvironment(t *testing.T) {
	base := &recordingRunner{}
	runner := nonInteractiveRunner{Runner: base}
	runner.Run(context.Background(), "proot-distro", "login", "ubuntu", "--", "env", "PATH=/usr/bin", "apt-get", "install")

	want := []string{"login", "ubuntu", "--", "env", "DEBIAN_FRONTEND=noninteractive", "PATH=/usr/bin", "apt-get", "install"}
	if base.name != "proot-distro" || !slices.Equal(base.args, want) {
		t.Fatalf("command = %s %v, want %s %v", base.name, base.args, "proot-distro", want)
	}
}

func TestRunnerForSelectsInteractiveRunner(t *testing.T) {
	if _, ok := runnerFor(Options{Interactive: true}).(InteractiveRunner); !ok {
		t.Fatal("interactive options did not select InteractiveRunner")
	}
	if _, ok := runnerFor(Options{}).(nonInteractiveRunner); !ok {
		t.Fatal("default options did not select non-interactive runner")
	}
}

func TestInteractiveRunnerCapturesVisibleCommandOutput(t *testing.T) {
	result := (InteractiveRunner{}).Run(context.Background(), "/bin/sh", "-ec", "printf 'hello\\n'")
	if result.Err != nil || !strings.Contains(string(result.Stdout), "hello") {
		t.Fatalf("interactive result = %+v, want successful hello output", result)
	}
}

func TestCopyTerminalInputStopsWhenCancelled(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		copyTerminalInput(ctx, io.Discard, reader)
		close(done)
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("terminal input forwarding did not stop after cancellation")
	}
}

func TestNonInteractiveRunnerPreservesRunnerErrors(t *testing.T) {
	base := errorRunner{err: errors.New("runner failed")}
	result := (nonInteractiveRunner{Runner: base}).Run(context.Background(), "apt-get", "install")
	if !errors.Is(result.Err, base.err) {
		t.Fatalf("error = %v, want %v", result.Err, base.err)
	}
}

type errorRunner struct{ err error }

func (r errorRunner) Run(context.Context, string, ...string) CommandResult {
	return CommandResult{Err: r.err}
}
