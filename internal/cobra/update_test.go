package cobra

import (
	"errors"
	"testing"

	"github.com/ericklucioh/mobdesk/internal/update"
)

func TestUpdateOperationResult(t *testing.T) {
	tests := []struct {
		name      string
		result    update.Result
		err       error
		checkOnly bool
		success   bool
		state     string
		message   string
	}{
		{
			name:    "already current",
			result:  update.Result{CurrentVersion: "v1.0.0", LatestVersion: "v1.0.0"},
			success: true,
			state:   "current",
			message: "Mobdesk v1.0.0 já está atualizado",
		},
		{
			name:      "update available",
			result:    update.Result{CurrentVersion: "v1.0.0", LatestVersion: "v1.1.0", Updated: true},
			checkOnly: true,
			success:   true,
			state:     "available",
			message:   "Atualização disponível: v1.0.0 → v1.1.0",
		},
		{
			name:    "updated",
			result:  update.Result{CurrentVersion: "v1.0.0", LatestVersion: "v1.1.0", Updated: true},
			success: true,
			state:   "updated",
			message: "Mobdesk atualizado: v1.0.0 → v1.1.0",
		},
		{
			name:    "failure",
			result:  update.Result{CurrentVersion: "v1.0.0"},
			err:     errors.New("rede indisponível"),
			success: false,
			state:   "failed",
			message: "rede indisponível",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := updateOperationResult(test.result, test.err, test.checkOnly)
			if response.Command != "update" || response.Success != test.success || response.State != test.state || response.Message != test.message {
				t.Fatalf("response = %+v", response)
			}
			if response.CurrentVersion != test.result.CurrentVersion || response.LatestVersion != test.result.LatestVersion || response.Updated != test.result.Updated {
				t.Fatalf("version fields were not preserved: %+v", response)
			}
		})
	}
}
