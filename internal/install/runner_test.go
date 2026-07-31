package install

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
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

	want := []string{"login", "ubuntu", "--", "env", "DEBIAN_FRONTEND=noninteractive", "TZ=Etc/UTC", "PATH=/usr/bin", "apt-get", "install"}
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
