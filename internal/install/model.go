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

// AppProfile is the domain model for the app catalog. Installation strategy
// details remain declarative data until the corresponding service phases.
type AppProfile struct {
	Name             string           `json:"name"`
	Aliases          []string         `json:"aliases"`
	Description      string           `json:"description,omitempty"`
	Kind             string           `json:"kind,omitempty"`
	Package          string           `json:"package"`
	Executable       string           `json:"executable"`
	VersionArg       []string         `json:"version_arg"`
	InstallKind      string           `json:"install_kind,omitempty"`
	Requires         []string         `json:"requires,omitempty"`
	UserBin          bool             `json:"-"`
	Script           string           `json:"-"`
	InstallProfile   string           `json:"install_profile,omitempty"`
	UninstallProfile string           `json:"uninstall_profile,omitempty"`
	ConfigProfile    string           `json:"config_profile,omitempty"`
	ConfigTarget     string           `json:"config_target,omitempty"`
	MinimumVersion   string           `json:"minimum_version,omitempty"`
	ProfileVersion   string           `json:"profile_version,omitempty"`
	StorageEstimate  *StorageEstimate `json:"storage_estimate,omitempty"`
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

func (s StorageEstimate) TotalMinMB() int64 {
	return s.AppMinMB + s.DependenciesMinMB + s.ConfigMinMB
}

func (s StorageEstimate) TotalMaxMB() int64 {
	return s.AppMaxMB + s.DependenciesMaxMB + s.ConfigMaxMB
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
