package cobra

import (
	"os"
	"strings"
	"testing"

	"github.com/ericklucioh/mobdesk/internal/paths"
)

func TestEnsureSetupCompletedRequiresMarker(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	err := ensureSetupCompleted(paths.New(os.Getenv("HOME"), ""))
	if err == nil || !strings.Contains(err.Error(), "mobdesk setup") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureSetupCompletedAcceptsMarker(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := paths.New(home, "").DataDir()
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.New(home, "").SetupDone(), []byte("setup concluido\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := ensureSetupCompleted(paths.New(home, "")); err != nil {
		t.Fatal(err)
	}
}
