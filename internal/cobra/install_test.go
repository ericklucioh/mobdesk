package cobra

import (
	"errors"
	"testing"

	"github.com/ericklucioh/mobdesk/internal/install"
)

func TestInstallOperationResultReportsFailureAsJSON(t *testing.T) {
	installErr := errors.New("apt-get falhou")
	result := installOperationResult(install.Result{Language: "go", State: "installing", LogPath: "/private/install.log"}, installErr)

	if result.Success || result.State != "failed" || result.Message != installErr.Error() || result.Language != "go" || result.LogPath != "/private/install.log" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestInstallOperationResultReportsSuccessAsJSON(t *testing.T) {
	result := installOperationResult(install.Result{Language: "go", Version: "go1.26", State: "installed"}, nil)

	if !result.Success || result.State != "installed" || result.Language != "go" || result.Version != "go1.26" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
