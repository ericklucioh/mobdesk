package install

import (
	"time"

	"github.com/ericklucioh/mobdesk/internal/i18n"
)

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
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
	// Description is localized presentation text; Usage is the concise command
	// form shown to users. Neither field controls installation.
	Description         string           `json:"description,omitempty"`
	Usage               string           `json:"usage,omitempty"`
	DescriptionID       i18n.MessageID   `json:"-"`
	Kind                string           `json:"kind,omitempty"`
	Package             string           `json:"package"`
	Packages            []string         `json:"packages,omitempty"`
	Executable          string           `json:"executable"`
	RequiredExecutables []ExecutableSpec `json:"required_executables,omitempty"`
	VersionArg          []string         `json:"version_arg"`
	// CatalogVersion is a short display fallback when VersionArg is not a
	// reliable, concise version source, such as an app's help command.
	CatalogVersion   string           `json:"catalog_version,omitempty"`
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

// ExecutableSpec describes one command that must be available for a profile
// to be considered installed. The legacy executable fields remain supported.
type ExecutableSpec struct {
	Name       string   `json:"name"`
	VersionArg []string `json:"version_arg,omitempty"`
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
	Packages        []string         `json:"packages,omitempty"`
	Executable      string           `json:"executable"`
	Executables     []ExecutableSpec `json:"executables,omitempty"`
	Version         string           `json:"version"`
	Installed       bool             `json:"installed"`
	Changed         bool             `json:"changed"`
	State           string           `json:"state"`
	LogPath         string           `json:"log_path"`
	Source          string           `json:"source,omitempty"`
	Paths           []string         `json:"paths,omitempty"`
	Conflicts       []string         `json:"conflicts,omitempty"`
	StorageEstimate *StorageEstimate `json:"storage_estimate,omitempty"`
}

type InstallationRecord struct {
	Name                string            `json:"name"`
	Kind                string            `json:"kind"`
	Package             string            `json:"package"`
	Packages            []string          `json:"packages,omitempty"`
	Executable          string            `json:"executable"`
	RequiredExecutables []ExecutableSpec  `json:"required_executables,omitempty"`
	Strategy            string            `json:"strategy,omitempty"`
	Dependencies        []string          `json:"dependencies,omitempty"`
	InstalledPackages   []string          `json:"installed_packages,omitempty"`
	InstalledFiles      []string          `json:"installed_files,omitempty"`
	InstalledDirs       []string          `json:"installed_directories,omitempty"`
	InstalledFileHashes map[string]string `json:"installed_file_hashes,omitempty"`
	RemovedPackages     []string          `json:"removed_packages,omitempty"`
	RemovedFiles        []string          `json:"removed_files,omitempty"`
	PreservedFiles      []string          `json:"preserved_files,omitempty"`
	State               string            `json:"state"`
	Source              string            `json:"source,omitempty"`
	Version             string            `json:"version,omitempty"`
	InstalledAt         time.Time         `json:"installed_at,omitempty"`
	LastAttemptAt       time.Time         `json:"last_attempt_at"`
	LastError           string            `json:"last_error,omitempty"`
	LastErrorCode       string            `json:"last_error_code,omitempty"`
	LogPath             string            `json:"log_path"`
}

type ConfigurationRecord struct {
	App            string            `json:"app"`
	Profile        string            `json:"profile"`
	ProfileVersion string            `json:"profile_version"`
	State          ConfigState       `json:"state"`
	ManagedPaths   []string          `json:"managed_paths,omitempty"`
	GeneratedFiles []string          `json:"generated_files,omitempty"`
	ManagedPlugins []string          `json:"managed_plugins,omitempty"`
	FileHashes     map[string]string `json:"file_hashes,omitempty"`
	AppliedAt      time.Time         `json:"applied_at,omitempty"`
	RemovedAt      time.Time         `json:"removed_at,omitempty"`
	ModifiedPaths  []string          `json:"modified_paths,omitempty"`
	Conflicts      []string          `json:"conflicts,omitempty"`
	LastError      string            `json:"last_error,omitempty"`
	LastErrorCode  string            `json:"last_error_code,omitempty"`
}

type ConfigFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Mode    uint32 `json:"mode,omitempty"`
}

type ConfigCommand struct {
	Name string   `json:"name"`
	Args []string `json:"args,omitempty"`
}

type ConfigProfile struct {
	ID              string           `json:"id"`
	Version         string           `json:"version"`
	App             string           `json:"app"`
	Description     string           `json:"description"`
	DescriptionID   i18n.MessageID   `json:"-"`
	ManagedPaths    []string         `json:"managed_paths"`
	Files           []ConfigFile     `json:"files"`
	Plugins         []ConfigPlugin   `json:"plugins,omitempty"`
	ManagedPlugins  []string         `json:"managed_plugins,omitempty"`
	Validation      []ConfigCommand  `json:"validation,omitempty"`
	ConflictPolicy  string           `json:"conflict_policy"`
	StorageEstimate *StorageEstimate `json:"storage_estimate,omitempty"`
}

// ConfigPlugin identifies a fixed revision cloned into a profile-owned path.
// Repository and commit are catalog data; they are never formed from input.
type ConfigPlugin struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Commit     string `json:"commit"`
	Path       string `json:"path"`
}

type ConfigOperationResult struct {
	SchemaVersion   int              `json:"schema_version"`
	Command         string           `json:"command"`
	App             string           `json:"app"`
	Action          string           `json:"action"`
	Success         bool             `json:"success"`
	State           ConfigState      `json:"state"`
	Changed         bool             `json:"changed"`
	Message         string           `json:"message"`
	MessageID       string           `json:"message_id,omitempty"`
	ErrorCode       string           `json:"error_code,omitempty"`
	Profile         string           `json:"profile,omitempty"`
	StorageEstimate *StorageEstimate `json:"storage_estimate,omitempty"`
	Conflicts       []string         `json:"conflicts,omitempty"`
	Paths           []string         `json:"paths,omitempty"`
}
