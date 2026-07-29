package cobra

import (
	"os"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/install"
)

type operationResult struct {
	SchemaVersion   int                      `json:"schema_version"`
	Command         string                   `json:"command"`
	Success         bool                     `json:"success"`
	State           string                   `json:"state"`
	Message         string                   `json:"message"`
	Locale          string                   `json:"locale,omitempty"`
	MessageID       string                   `json:"message_id,omitempty"`
	ErrorCode       string                   `json:"error_code,omitempty"`
	Target          string                   `json:"target,omitempty"`
	Action          string                   `json:"action,omitempty"`
	Changed         bool                     `json:"changed"`
	ConfigState     string                   `json:"config_state,omitempty"`
	Conflicts       []string                 `json:"conflicts,omitempty"`
	Paths           []string                 `json:"paths,omitempty"`
	Source          string                   `json:"source,omitempty"`
	CurrentVersion  string                   `json:"current_version,omitempty"`
	LatestVersion   string                   `json:"latest_version,omitempty"`
	Updated         bool                     `json:"updated,omitempty"`
	LogPath         string                   `json:"log_path,omitempty"`
	Language        string                   `json:"language,omitempty"`
	Version         string                   `json:"version,omitempty"`
	Port            int                      `json:"port,omitempty"`
	Addresses       []string                 `json:"addresses,omitempty"`
	StorageEstimate *install.StorageEstimate `json:"storage_estimate,omitempty"`
}

func decorateResult(result operationResult, localizers []i18n.Localizer, messageID i18n.MessageID, operationErr error) operationResult {
	if len(localizers) == 0 {
		return result
	}
	result.Locale = string(localizers[0].Locale)
	result.MessageID = string(messageID)
	if operationErr != nil {
		result.ErrorCode = "operation_failed"
	}
	return result
}

func withQuietOutput(run func() error) error {
	previous := os.Stdout
	previousErr := os.Stderr
	quiet, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	os.Stdout = quiet
	os.Stderr = quiet
	runErr := run()
	os.Stdout = previous
	os.Stderr = previousErr
	_ = quiet.Close()
	return runErr
}
