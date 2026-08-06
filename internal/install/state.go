package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/paths"
)

func SaveConfigurationRecord(p paths.Paths, record ConfigurationRecord) error {
	if err := validateStateName(record.App); err != nil {
		return err
	}
	return writePrivateJSON(p.ConfigurationState(record.App), record)
}

func LoadConfigurationRecord(p paths.Paths, app string) (ConfigurationRecord, error) {
	if err := validateStateName(app); err != nil {
		return ConfigurationRecord{}, err
	}
	payload, err := os.ReadFile(p.ConfigurationState(app))
	if err != nil {
		return ConfigurationRecord{}, err
	}
	var record ConfigurationRecord
	if err := json.Unmarshal(payload, &record); err != nil {
		return ConfigurationRecord{}, i18n.NewError(i18n.ServiceConfigError, "config_read_state", map[string]any{"Detail": app}, err)
	}
	return record, nil
}

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

func writePrivateJSON(path string, value any) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".state-*.tmp")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(append(payload, '\n')); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryName, path)
}

func validateStateName(name string) error {
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) {
		return i18n.NewError(i18n.ServiceInstallState, "state_name_invalid", map[string]any{"Detail": name}, nil)
	}
	return nil
}
