package install

import "time"

type Runtime string

const (
	RuntimeTermux Runtime = "termux"
	RuntimeUbuntu Runtime = "ubuntu"
)

// Target describes the package-manager strategy for one runtime.
type Target struct {
	Runtime     Runtime `json:"runtime"`
	Package     string  `json:"package"`
	InstallKind string  `json:"install_kind"`
}

type Language struct {
	Name        string   `json:"name"`
	Aliases     []string `json:"aliases"`
	Package     string   `json:"package"`
	Executable  string   `json:"executable"`
	VersionArg  []string `json:"version_arg"`
	Kind        string   `json:"kind,omitempty"`
	InstallKind string   `json:"install_kind,omitempty"`
	Requires    []string `json:"requires,omitempty"`
	Script      string   `json:"-"`
	UserBin     bool     `json:"-"`
	Targets     []Target `json:"targets,omitempty"`
}

func (l Language) Target(runtime Runtime) (Target, bool) {
	for _, target := range l.Targets {
		if target.Runtime == runtime {
			return target, true
		}
	}
	return Target{}, false
}

type Result struct {
	SchemaVersion int     `json:"schema_version"`
	Language      string  `json:"language"`
	Package       string  `json:"package"`
	Executable    string  `json:"executable"`
	Version       string  `json:"version"`
	Installed     bool    `json:"installed"`
	Changed       bool    `json:"changed"`
	State         string  `json:"state"`
	LogPath       string  `json:"log_path"`
	Runtime       Runtime `json:"runtime"`
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
	Runtime       Runtime   `json:"runtime"`
}
