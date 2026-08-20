package status

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/ericklucioh/mobdesk/internal/paths"
)

type javaStatusRunner struct{}

func (javaStatusRunner) Run(context.Context, string, ...string) CommandResult {
	return CommandResult{Stdout: []byte("openjdk 21.0.12\n")}
}

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

func TestCollectReportsManagedJavaStatus(t *testing.T) {
	prefix := "/data/data/com.termux/files/usr"
	p := paths.New(t.TempDir(), prefix)
	if err := os.MkdirAll(p.InstallationsDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	record := install.InstallationRecord{Name: "java", Kind: "language", Package: "openjdk-21", Executable: "java", RequiredExecutables: []install.ExecutableSpec{{Name: "java", VersionArg: []string{"--version"}}, {Name: "javac", VersionArg: []string{"--version"}}, {Name: "jar", VersionArg: []string{"--version"}}}, State: "installed", Source: "mobdesk", Version: "openjdk 21.0.12", JavaHome: prefix + "/lib/jvm/java-21-openjdk"}
	payload, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p.InstallationsDir()+"/java.json", payload, 0o600); err != nil {
		t.Fatal(err)
	}
	value := Collect(context.Background(), Options{Paths: p, CommandRunner: javaStatusRunner{}, LookPath: func(string) (string, error) { return "", os.ErrNotExist }})
	if !value.Java.Installed || value.Java.State != CheckOK || value.Java.Home != record.JavaHome || value.Java.Version != record.Version {
		t.Fatalf("unexpected Java status: %+v", value.Java)
	}
	payload, err = json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil || raw["java"] == nil {
		t.Fatalf("Java status is not present in JSON: %v", err)
	}
}

func TestCollectJavaTreatsUninstalledProfileAsOptional(t *testing.T) {
	value := collectJava([]InstallationStatus{{Name: "java", State: "uninstalled"}}, "/data/data/com.termux/files/usr")
	if value.State != CheckMissing || value.Installed || value.Error != "java_not_installed" {
		t.Fatalf("uninstalled Java status = %+v", value)
	}
}
