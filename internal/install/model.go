package install

import "time"

type Language struct {
	Name       string   `json:"name"`
	Aliases    []string `json:"aliases"`
	Package    string   `json:"package"`
	Executable string   `json:"executable"`
	VersionArg []string `json:"version_arg"`
}

type Result struct {
	Language   string `json:"language"`
	Package    string `json:"package"`
	Executable string `json:"executable"`
	Version    string `json:"version"`
	Installed  bool   `json:"installed"`
	Changed    bool   `json:"changed"`
	State      string `json:"state"`
	LogPath    string `json:"log_path"`
}

type InstallationRecord struct {
	Name          string    `json:"name"`
	Kind          string    `json:"kind"`
	Package       string    `json:"package"`
	Executable    string    `json:"executable"`
	State         string    `json:"state"`
	Version       string    `json:"version,omitempty"`
	InstalledAt   time.Time `json:"installed_at,omitempty"`
	LastAttemptAt time.Time `json:"last_attempt_at"`
	LastError     string    `json:"last_error,omitempty"`
	LogPath       string    `json:"log_path"`
}
