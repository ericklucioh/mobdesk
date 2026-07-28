package logs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ericklucioh/mobdesk/internal/paths"
)

func TestReadReturnsRecentInstallationLogLines(t *testing.T) {
	home := t.TempDir()
	p := paths.New(home, "")
	stateDir := p.InstallationsDir()
	logPath := filepath.Join(p.InstallLogsDir(), "go.log")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p.InstallLogsDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "go.json"), []byte(`{"name":"go","kind":"language","state":"installed","version":"1.26","log_path":"`+logPath+`"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logPath, []byte("linha 1\nlinha 2\nlinha 3\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	snapshot, err := Read(Options{Paths: p, Lines: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Logs) != 1 || snapshot.Logs[0].Name != "go" {
		t.Fatalf("unexpected logs: %+v", snapshot.Logs)
	}
	if snapshot.Logs[0].Missing || snapshot.Logs[0].Content != "linha 2\nlinha 3" {
		t.Fatalf("unexpected tail: %+v", snapshot.Logs[0])
	}
}

func TestReadFiltersByNameAndMarksMissingFile(t *testing.T) {
	home := t.TempDir()
	p := paths.New(home, "")
	stateDir := p.InstallationsDir()
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	payload := `{"name":"python","kind":"language","state":"failed","log_path":"/missing/python.log"}`
	if err := os.WriteFile(filepath.Join(stateDir, "python.json"), []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}

	snapshot, err := Read(Options{Paths: p, Name: "PYTHON"})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Logs) != 1 || !snapshot.Logs[0].Missing {
		t.Fatalf("missing log was not reported: %+v", snapshot.Logs)
	}
	if !strings.EqualFold(snapshot.Logs[0].Name, "python") {
		t.Fatalf("unexpected filtered log: %+v", snapshot.Logs[0])
	}
}

func TestReadIgnoresPersistedLogPathOutsideMobdesk(t *testing.T) {
	home := t.TempDir()
	p := paths.New(home, "")
	if err := os.MkdirAll(p.InstallationsDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p.InstallLogsDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	sensitivePath := filepath.Join(home, "sensitive.txt")
	if err := os.WriteFile(sensitivePath, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	expectedPath := filepath.Join(p.InstallLogsDir(), "go.log")
	if err := os.WriteFile(expectedPath, []byte("safe log\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	payload := `{"name":"go","kind":"language","state":"installed","log_path":"` + sensitivePath + `"}`
	if err := os.WriteFile(filepath.Join(p.InstallationsDir(), "go.json"), []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}

	snapshot, err := Read(Options{Paths: p})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Logs) != 1 {
		t.Fatalf("unexpected logs: %+v", snapshot.Logs)
	}
	if snapshot.Logs[0].LogPath != expectedPath || snapshot.Logs[0].Content != "safe log" {
		t.Fatalf("persisted log path was trusted: %+v", snapshot.Logs[0])
	}
}
