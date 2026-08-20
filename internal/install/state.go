package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/paths"
)

func loadInstallationRecord(p paths.Paths, app string) (InstallationRecord, error) {
	if err := validateStateName(app); err != nil {
		return InstallationRecord{}, err
	}

	payload, err := os.ReadFile(filepath.Join(p.InstallationsDir(), app+".json"))
	if err != nil {
		return InstallationRecord{}, err
	}

	var record InstallationRecord
	if err := json.Unmarshal(payload, &record); err != nil {
		return InstallationRecord{}, i18n.NewError(i18n.ServiceInstallState, "install_read_state", map[string]any{"Detail": app}, err)
	}

	return record, nil
}

func validateStateName(name string) error {
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) {
		return i18n.NewError(i18n.ServiceInstallState, "state_name_invalid", map[string]any{"Detail": name}, nil)
	}
	return nil
}
