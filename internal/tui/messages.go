package tui

import (
	"encoding/json"

	tea "charm.land/bubbletea/v2"
	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/ericklucioh/mobdesk/internal/status"
	"github.com/ericklucioh/mobdesk/internal/version"
)

type operationMessage struct {
	id      int
	command string
	result  operationResult
	err     error
}

type operationProgressMessage struct {
	id      int
	message string
	next    tea.Cmd
}

type operationResult struct {
	Success         bool                     `json:"success"`
	State           string                   `json:"state"`
	Message         string                   `json:"message"`
	CurrentVersion  string                   `json:"current_version"`
	LatestVersion   string                   `json:"latest_version"`
	Updated         bool                     `json:"updated"`
	LogPath         string                   `json:"log_path"`
	Language        string                   `json:"language"`
	Version         string                   `json:"version"`
	Port            int                      `json:"port"`
	Addresses       []string                 `json:"addresses"`
	StorageEstimate *install.StorageEstimate `json:"storage_estimate,omitempty"`
}

type statusMessage struct {
	id    int
	value status.SystemStatus
	info  version.Info
	err   error
}

func decodeOperation(output []byte) (operationResult, error) {
	result := operationResult{}
	if err := json.Unmarshal(output, &result); err != nil {
		return operationResult{}, err
	}
	return result, nil
}
