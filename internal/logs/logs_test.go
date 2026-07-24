package logs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadReturnsRecentInstallationLogLines(t *testing.T) {
	home := t.TempDir()
	stateDir := filepath.Join(home, ".local", "share", "mobdesk", "state", "installations")
	logPath := filepath.Join(home, "go.log")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "go.json"), []byte(`{"name":"go","kind":"language","state":"installed","version":"1.26","log_path":"`+logPath+`"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logPath, []byte("linha 1\nlinha 2\nlinha 3\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	snapshot, err := Read(Options{Home: home, Lines: 2})
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
	stateDir := filepath.Join(home, ".local", "share", "mobdesk", "state", "installations")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	payload := `{"name":"python","kind":"language","state":"failed","log_path":"/missing/python.log"}`
	if err := os.WriteFile(filepath.Join(stateDir, "python.json"), []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}

	snapshot, err := Read(Options{Home: home, Name: "PYTHON"})
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
