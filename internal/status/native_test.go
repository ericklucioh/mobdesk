package status

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ericklucioh/mobdesk/internal/paths"
)

func TestCollectReportsNativeWorkspaceSchema(t *testing.T) {
	t.Setenv("TERMUX_VERSION", "1")
	p := paths.New(t.TempDir(), "/data/data/com.termux/files/usr")
	if err := os.MkdirAll(p.Workspace(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p.StateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"directories", "packages-updated", "packages-installed", "workspace-created", "password-configured", "ssh-configured", "shell-configured", "launcher-installed"} {
		if err := os.WriteFile(p.SetupPhase(phase), []byte("done"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(p.SetupDone(), []byte("done"), 0o600); err != nil {
		t.Fatal(err)
	}
	value := Collect(context.Background(), Options{Paths: p, Now: func() time.Time { return time.Unix(0, 0) }, LookPath: func(string) (string, error) { return "", os.ErrNotExist }})
	if value.SchemaVersion != 2 || !value.Workspace.Exists || value.Workspace.Path != p.Workspace() {
		t.Fatalf("unexpected native status: %+v", value)
	}
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["workspace"]; !ok {
		t.Fatal("status has no workspace")
	}
	if _, ok := raw["ubuntu"]; ok {
		t.Fatal("status retained a guest runtime")
	}
}
