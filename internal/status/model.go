package status

import (
	"time"

	"github.com/ericklucioh/mobdesk/internal/install"
)

type OverallState string

const (
	StateHealthy  OverallState = "healthy"
	StateDegraded OverallState = "degraded"
	StateError    OverallState = "error"
	StateUnknown  OverallState = "unknown"
)

type CheckState string

const (
	CheckOK      CheckState = "ok"
	CheckWarning CheckState = "warning"
	CheckError   CheckState = "error"
	CheckMissing CheckState = "missing"
	CheckUnknown CheckState = "unknown"
)

type SystemStatus struct {
	SchemaVersion int                  `json:"schema_version"`
	GeneratedAt   time.Time            `json:"generated_at"`
	Overall       OverallState         `json:"overall"`
	Host          HostStatus           `json:"host"`
	Setup         SetupStatus          `json:"setup"`
	Workspace     WorkspaceStatus      `json:"workspace"`
	Storage       StorageStatus        `json:"storage"`
	SSH           SSHStatus            `json:"ssh"`
	Network       NetworkStatus        `json:"network"`
	Battery       BatteryStatus        `json:"battery"`
	WiFi          WiFiStatus           `json:"wifi"`
	Installations []InstallationStatus `json:"installations"`
	Alerts        AlertSummary         `json:"alerts"`
}

type HostStatus struct {
	State              CheckState `json:"state"`
	Termux             bool       `json:"termux"`
	OS                 string     `json:"os"`
	Architecture       string     `json:"architecture"`
	Home               string     `json:"home"`
	Prefix             string     `json:"prefix"`
	OpenSSH            bool       `json:"openssh"`
	Ifconfig           bool       `json:"ifconfig"`
	WakeLockAvailable  bool       `json:"wake_lock_available"`
	TermuxAPIAvailable bool       `json:"termux_api_available"`
	Error              string     `json:"error,omitempty"`
}

type SetupStatus struct {
	State     CheckState        `json:"state"`
	Completed bool              `json:"completed"`
	Phases    map[string]string `json:"phases"`
}

type WorkspaceStatus struct {
	State  CheckState `json:"state"`
	Path   string     `json:"path"`
	Exists bool       `json:"exists"`
	Error  string     `json:"error,omitempty"`
}

type StorageStatus struct {
	State       CheckState `json:"state"`
	Warning     bool       `json:"warning,omitempty"`
	Blocked     bool       `json:"blocked,omitempty"`
	DeviceTotal int64      `json:"device_total_bytes"`
	DeviceUsed  int64      `json:"device_used_bytes"`
	DeviceFree  int64      `json:"device_free_bytes"`
	Error       string     `json:"error,omitempty"`
}

type SSHStatus struct {
	State        CheckState `json:"state"`
	Enabled      bool       `json:"enabled"`
	Running      bool       `json:"running"`
	Port         int        `json:"port"`
	PID          int        `json:"pid,omitempty"`
	ConfigPath   string     `json:"config_path"`
	LogPath      string     `json:"log_path"`
	ConfigExists bool       `json:"config_exists"`
	Error        string     `json:"error,omitempty"`
}

type NetworkStatus struct {
	State     CheckState `json:"state"`
	Addresses []string   `json:"addresses"`
	Preferred string     `json:"preferred,omitempty"`
	Error     string     `json:"error,omitempty"`
}

type BatteryStatus struct {
	State       CheckState `json:"state"`
	Available   bool       `json:"available"`
	Percentage  *int       `json:"percentage,omitempty"`
	Status      string     `json:"status,omitempty"`
	Plugged     string     `json:"plugged,omitempty"`
	Temperature *float64   `json:"temperature,omitempty"`
	Health      string     `json:"health,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type WiFiStatus struct {
	State         CheckState `json:"state"`
	Available     bool       `json:"available"`
	Connected     bool       `json:"connected"`
	SSID          string     `json:"ssid,omitempty"`
	IP            string     `json:"ip,omitempty"`
	LinkSpeedMbps *int       `json:"link_speed_mbps,omitempty"`
	FrequencyMHz  *int       `json:"frequency_mhz,omitempty"`
	Error         string     `json:"error,omitempty"`
}

type InstallationStatus struct {
	Name                string                   `json:"name"`
	Kind                string                   `json:"kind"`
	Package             string                   `json:"package"`
	Packages            []string                 `json:"packages,omitempty"`
	Executable          string                   `json:"executable"`
	RequiredExecutables []install.ExecutableSpec `json:"required_executables,omitempty"`
	MissingExecutables  []string                 `json:"missing_executables,omitempty"`
	MissingDependencies []string                 `json:"missing_dependencies,omitempty"`
	Strategy            string                   `json:"strategy,omitempty"`
	Dependencies        []string                 `json:"dependencies,omitempty"`
	InstalledPackages   []string                 `json:"installed_packages,omitempty"`
	State               string                   `json:"state"`
	Source              string                   `json:"source,omitempty"`
	Managed             bool                     `json:"managed,omitempty"`
	Version             string                   `json:"version,omitempty"`
	InstalledAt         time.Time                `json:"installed_at,omitempty"`
	LastAttemptAt       time.Time                `json:"last_attempt_at"`
	LastError           string                   `json:"last_error,omitempty"`
	LastErrorCode       string                   `json:"last_error_code,omitempty"`
	LogPath             string                   `json:"log_path"`
	StorageEstimate     *install.StorageEstimate `json:"storage_estimate,omitempty"`
}

type AlertSummary struct {
	OK       int `json:"ok"`
	Warnings int `json:"warnings"`
	Errors   int `json:"errors"`
	Missing  int `json:"missing"`
	Unknown  int `json:"unknown"`
}
