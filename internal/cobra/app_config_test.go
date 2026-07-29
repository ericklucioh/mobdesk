package cobra

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/spf13/cobra"
)

func TestConfigOperationResultIncludesConfigContract(t *testing.T) {
	result := configOperationResult(install.ConfigOperationResult{
		SchemaVersion: 1,
		App:           "neovim",
		Action:        "apply",
		Success:       true,
		State:         install.ConfigStateApplied,
		Changed:       true,
		Message:       "Configuração aplicada",
		Paths:         []string{"/root/.config/nvim"},
		StorageEstimate: &install.StorageEstimate{
			ConfigMinMB: 100,
			ConfigMaxMB: 300,
		},
	}, nil)
	if !result.Success || result.Target != "neovim" || result.Action != "apply" || result.ConfigState != "applied" || !result.Changed {
		t.Fatalf("unexpected config result: %+v", result)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), `"storage_estimate"`) || !strings.Contains(string(payload), `"paths"`) {
		t.Fatalf("config contract omitted additive fields: %s", payload)
	}
}

func TestConfigEmitsValidJSONWhenRuntimeIsBlocked(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PREFIX", "/not-termux")
	t.Setenv("TERMUX_VERSION", "")
	previous := configApplyJSON
	configApplyJSON = true
	defer func() { configApplyJSON = previous }()

	output, err := captureStdout(func() error { return runConfigOperation(context.Background(), "neovim", "apply", true, false) })
	if err == nil || !strings.Contains(err.Error(), "deve ser executado no Termux") {
		t.Fatalf("unexpected runtime error: %v", err)
	}
	var result operationResult
	if decodeErr := json.Unmarshal([]byte(output), &result); decodeErr != nil {
		t.Fatalf("runtime output is not JSON: %q: %v", output, decodeErr)
	}
	if result.Success || result.Target != "neovim" || result.Action != "apply" || result.ConfigState != "failed" {
		t.Fatalf("unexpected runtime JSON: %+v", result)
	}
}

func TestConfigCommandsRequireExactlyOneApp(t *testing.T) {
	for _, command := range []*cobra.Command{configApplyCmd, configRemoveCmd} {
		if err := command.Args(command, nil); err == nil {
			t.Fatalf("%s accepted no app", command.Use)
		}
		if err := command.Args(command, []string{"neovim", "extra"}); err == nil {
			t.Fatalf("%s accepted multiple apps", command.Use)
		}
	}
}
