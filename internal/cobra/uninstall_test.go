package cobra

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ericklucioh/mobdesk/internal/install"
)

func TestUninstallOperationResultKeepsStructuredFailure(t *testing.T) {
	result := uninstallOperationResult(install.Result{
		Language: "neovim", State: "failed", Source: "detected",
		Conflicts: []string{"/root/.local/bin/nvim"},
	}, "nvim", errors.New("instalação apenas detectada"))
	if result.Success || result.Target != "nvim" || result.Action != "uninstall" || result.State != "failed" || result.Source != "detected" {
		t.Fatalf("unexpected uninstall result: %+v", result)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var decoded operationResult
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.SchemaVersion != 1 || decoded.Message == "" || len(decoded.Conflicts) != 1 {
		t.Fatalf("structured uninstall failure lost fields: %+v", decoded)
	}
}

func TestUninstallEmitsValidJSONWhenRuntimeIsBlocked(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PREFIX", "/not-termux")
	t.Setenv("TERMUX_VERSION", "")
	previous := uninstallJSON
	uninstallJSON = true
	defer func() { uninstallJSON = previous }()

	output, err := captureStdout(func() error { return runUninstall(context.Background(), "neovim") })
	if err == nil || !strings.Contains(err.Error(), "deve ser executado no Termux") {
		t.Fatalf("unexpected runtime error: %v", err)
	}
	var result operationResult
	if decodeErr := json.Unmarshal([]byte(output), &result); decodeErr != nil {
		t.Fatalf("runtime output is not JSON: %q: %v", output, decodeErr)
	}
	if result.Success || result.Target != "neovim" || result.Action != "uninstall" || result.State != "failed" {
		t.Fatalf("unexpected runtime JSON: %+v", result)
	}
}

func TestUninstallCommandRequiresExactlyOneApp(t *testing.T) {
	if err := uninstallCmd.Args(uninstallCmd, nil); err == nil {
		t.Fatal("uninstall accepted no app")
	}
	if err := uninstallCmd.Args(uninstallCmd, []string{"neovim", "extra"}); err == nil {
		t.Fatal("uninstall accepted multiple apps")
	}
}

func captureStdout(run func() error) (string, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return "", err
	}
	previous := os.Stdout
	os.Stdout = write
	runErr := run()
	os.Stdout = previous
	_ = write.Close()
	payload, readErr := io.ReadAll(read)
	_ = read.Close()
	if runErr != nil {
		return string(payload), runErr
	}
	return string(payload), readErr
}
