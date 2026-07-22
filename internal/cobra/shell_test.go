package cobra

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureSetupCompletedRequiresMarker(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	err := ensureSetupCompleted()
	if err == nil || !strings.Contains(err.Error(), "mobdesk setup") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureSetupCompletedAcceptsMarker(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".local", "share", "mobdesk")
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "setup.done"), []byte("setup concluido\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := ensureSetupCompleted(); err != nil {
		t.Fatal(err)
	}
}
