package cobra

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/install"
)

func TestInstallOperationResultReportsFailureAsJSON(t *testing.T) {
	installErr := errors.New("apt-get failed")
	result := installOperationResult(install.Result{Language: "go", State: "installing", LogPath: "/private/install.log"}, installErr)

	wantMessage := i18n.New(i18n.LocaleENUS).Text(i18n.ErrorOperationFailed, map[string]any{"Detail": installErr.Error()})
	if result.Success || result.State != "failed" || result.Message != wantMessage || result.Language != "go" || result.LogPath != "/private/install.log" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestInstallOperationResultReportsSuccessAsJSON(t *testing.T) {
	result := installOperationResult(install.Result{Language: "go", Version: "go1.26", State: "installed"}, nil)

	if !result.Success || result.State != "installed" || result.Language != "go" || result.Version != "go1.26" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestInstallOperationResultPreservesStorageEstimate(t *testing.T) {
	estimate := &install.StorageEstimate{AppMinMB: 15, AppMaxMB: 30}
	result := installOperationResult(install.Result{Language: "neovim", StorageEstimate: estimate}, nil)
	if result.StorageEstimate != estimate {
		t.Fatalf("storage estimate was not preserved: %+v", result.StorageEstimate)
	}
}

func TestOperationResultKeepsStorageEstimateOptional(t *testing.T) {
	withoutEstimate, err := json.Marshal(operationResult{SchemaVersion: 1, Command: "install", Success: true, State: "installed", Message: "ok"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(withoutEstimate), "storage_estimate") {
		t.Fatalf("empty storage estimate changed the schema: %s", withoutEstimate)
	}

	withEstimate, err := json.Marshal(operationResult{
		SchemaVersion:   1,
		Command:         "install",
		Success:         true,
		State:           "installed",
		Message:         "ok",
		StorageEstimate: &install.StorageEstimate{AppMinMB: 20},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(withEstimate), "storage_estimate") {
		t.Fatalf("storage estimate was not emitted: %s", withEstimate)
	}
}

func TestInstallRejectsUbuntuRuntime(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PREFIX", "/not-termux")
	t.Setenv("TERMUX_VERSION", "")

	err := runInstall(context.Background(), "zellij")
	wantError := i18n.New(i18n.LocaleENUS).Text(i18n.ErrorTermuxRequired, map[string]any{"Operation": "mobdesk install"})
	if err == nil || err.Error() != wantError {
		t.Fatalf("unexpected error: %v", err)
	}
}
