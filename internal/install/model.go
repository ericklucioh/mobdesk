package install

import "time"

type AppState string

const (
	AppStateAvailable    AppState = "available"
	AppStateInstalling   AppState = "installing"
	AppStateInstalled    AppState = "installed"
	AppStateUninstalling AppState = "uninstalling"
	AppStateUninstalled  AppState = "uninstalled"
	AppStatePartial      AppState = "partial"
	AppStateFailed       AppState = "failed"
)

type ConfigState string

const (
	ConfigStateUnavailable ConfigState = "unavailable"
	ConfigStateNotApplied  ConfigState = "not_applied"
	ConfigStateApplying    ConfigState = "applying"
	ConfigStateApplied     ConfigState = "applied"
	ConfigStateRemoving    ConfigState = "removing"
	ConfigStateRemoved     ConfigState = "removed"
	ConfigStateModified    ConfigState = "modified"
	ConfigStateConflict    ConfigState = "conflict"
	ConfigStateFailed      ConfigState = "failed"
)

// AppProfile is the additive domain model for the app catalog. The existing
// Language catalog remains in use until the catalog migration phase.
type AppProfile struct {
	Name            string           `json:"name"`
	Aliases         []string         `json:"aliases"`
	Description     string           `json:"description,omitempty"`
	Kind            string           `json:"kind,omitempty"`
	Package         string           `json:"package"`
	Executable      string           `json:"executable"`
	VersionArg      []string         `json:"version_arg"`
	InstallKind     string           `json:"install_kind,omitempty"`
	Requires        []string         `json:"requires,omitempty"`
	UserBin         bool             `json:"user_bin,omitempty"`
	StorageEstimate *StorageEstimate `json:"storage_estimate,omitempty"`
}

type StorageEstimate struct {
	AppMinMB          int64     `json:"app_min_mb"`
	AppMaxMB          int64     `json:"app_max_mb"`
	DependenciesMinMB int64     `json:"dependencies_min_mb"`
	DependenciesMaxMB int64     `json:"dependencies_max_mb"`
	ConfigMinMB       int64     `json:"config_min_mb"`
	ConfigMaxMB       int64     `json:"config_max_mb"`
	Source            string    `json:"source"`
	Version           string    `json:"version"`
	Architecture      string    `json:"architecture"`
	MeasuredAt        time.Time `json:"measured_at"`
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
}

type Result struct {
	SchemaVersion   int              `json:"schema_version"`
	Language        string           `json:"language"`
	Package         string           `json:"package"`
	Executable      string           `json:"executable"`
	Version         string           `json:"version"`
	Installed       bool             `json:"installed"`
	Changed         bool             `json:"changed"`
	State           string           `json:"state"`
	LogPath         string           `json:"log_path"`
	StorageEstimate *StorageEstimate `json:"storage_estimate,omitempty"`
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
