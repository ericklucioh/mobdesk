package status

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ericklucioh/mobdesk/internal/executil"
	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/ericklucioh/mobdesk/internal/paths"
)

const (
	SchemaVersion  = 2
	SSHPort        = 8022
	commandTimeout = 2 * time.Second
	termuxTimeout  = time.Second
)

type Options struct {
	Paths         paths.Paths
	CommandRunner CommandRunner
	LookPath      func(string) (string, error)
	Now           func() time.Time
	SSHPort       int
	termux        bool
}

func (o Options) withDefaults() Options {
	o.termux = detectTermuxRuntime(o.Paths.Prefix)
	if o.CommandRunner == nil {
		o.CommandRunner = ExecRunner{}
	}
	if o.LookPath == nil {
		o.LookPath = executil.Resolve
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.SSHPort == 0 {
		o.SSHPort = SSHPort
	}
	return o
}

func Collect(ctx context.Context, options Options) SystemStatus {
	o := options.withDefaults()
	result := SystemStatus{SchemaVersion: SchemaVersion, GeneratedAt: o.Now().UTC(), Host: collectHost(o), Setup: collectSetup(o), Workspace: collectWorkspace(o), Storage: collectStorage(o), SSH: collectSSH(ctx, o), Network: collectNetwork(ctx, o)}
	result.Battery, result.WiFi = collectTermuxAPIs(ctx, o)
	result.Installations = reconcileInstallations(ctx, o, collectInstallations(o))
	result.Java = collectJava(result.Installations, o.Paths.Prefix)
	result.Alerts = summarize(result)
	result.Overall = overallState(result)
	return result
}

func collectJava(installations []InstallationStatus, prefix string) JavaStatus {
	result := JavaStatus{State: CheckMissing, Error: "java_not_installed"}
	for _, installation := range installations {
		if installation.Name != "java" {
			continue
		}
		if installation.State == "uninstalled" {
			return result
		}
		result.Installed, result.Version = installation.State == "installed", installation.Version
		if installation.State != "installed" {
			result.State, result.Error = CheckWarning, "java_installation_incomplete"
			return result
		}
		home, err := validJavaHome(installation.JavaHome, prefix)
		if err != nil {
			result.State, result.Error = CheckWarning, "java_home_invalid"
			return result
		}
		result.State, result.Home, result.Error = CheckOK, home, ""
		return result
	}
	return result
}

func validJavaHome(home, prefix string) (string, error) {
	home, prefix = filepath.Clean(strings.TrimSpace(home)), filepath.Clean(strings.TrimSpace(prefix))
	if !filepath.IsAbs(home) || !filepath.IsAbs(prefix) {
		return "", fmt.Errorf("java home or Termux prefix is not absolute")
	}
	relative, err := filepath.Rel(prefix, home)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("java home is outside Termux prefix")
	}
	return home, nil
}

func ReadInstallations(p paths.Paths) []InstallationStatus {
	return enrichInstallationMetadata(collectInstallations(Options{Paths: p}.withDefaults()))
}

func IsTermuxRuntime(prefix string) bool { return detectTermuxRuntime(prefix) }

func collectHost(o Options) HostStatus {
	result := HostStatus{State: CheckOK, Termux: o.termux, OS: runtime.GOOS, Architecture: runtime.GOARCH, Home: o.Paths.Home, Prefix: o.Paths.Prefix}
	result.OpenSSH = commandAvailable(o, "sshd")
	result.Ifconfig = commandAvailable(o, "ifconfig")
	result.WakeLockAvailable = commandAvailable(o, "termux-wake-lock")
	result.TermuxAPIAvailable = commandAvailable(o, "termux-battery-status") || commandAvailable(o, "termux-wifi-connectioninfo")
	if !result.Termux {
		result.State, result.Error = CheckMissing, "termux_runtime_unavailable"
	}
	return result
}

func collectSetup(o Options) SetupStatus {
	phases := []string{"directories", "packages-updated", "system-upgraded", "packages-installed", "workspace-created", "password-configured", "ssh-configured", "shell-configured", "launcher-installed"}
	required := map[string]bool{"directories": true, "packages-updated": true, "packages-installed": true, "workspace-created": true, "password-configured": true, "ssh-configured": true, "shell-configured": true, "launcher-installed": true}
	result := SetupStatus{State: CheckWarning, Phases: make(map[string]string, len(phases))}
	complete := true
	for _, phase := range phases {
		if _, err := os.Stat(o.Paths.SetupPhase(phase)); err == nil {
			result.Phases[phase] = "done"
		} else {
			result.Phases[phase] = "pending"
			complete = complete && !required[phase]
		}
	}
	result.Completed = complete && fileExists(o.Paths.SetupDone())
	if result.Completed {
		result.State = CheckOK
	}
	return result
}

func collectWorkspace(o Options) WorkspaceStatus {
	result := WorkspaceStatus{Path: o.Paths.Workspace(), State: CheckWarning}
	info, err := os.Stat(result.Path)
	if err == nil && info.IsDir() {
		result.State, result.Exists = CheckOK, true
		return result
	}
	if err != nil && !os.IsNotExist(err) {
		result.Error = "workspace_unavailable"
	} else {
		result.Error = "workspace_missing"
	}
	return result
}

func collectStorage(o Options) StorageStatus {
	result := StorageStatus{State: CheckUnknown}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(o.Paths.Home, &stat); err != nil {
		result.Error = "device_storage_unavailable"
		return result
	}
	blockSize := int64(stat.Bsize)
	result.DeviceTotal, result.DeviceFree = int64(stat.Blocks)*blockSize, int64(stat.Bavail)*blockSize
	result.DeviceUsed, result.State = result.DeviceTotal-int64(stat.Bfree)*blockSize, CheckOK
	result.Warning, result.Blocked = result.DeviceFree < install.StorageWarningBytes, result.DeviceFree < install.StorageBlockBytes
	return result
}

func collectSSH(ctx context.Context, o Options) SSHStatus {
	result := SSHStatus{State: CheckUnknown, Port: o.SSHPort, ConfigPath: o.Paths.SSHConfig(), LogPath: o.Paths.SSHLog(), ConfigExists: fileExists(o.Paths.SSHConfig())}
	if !o.termux {
		result.State, result.Error = CheckMissing, "termux_runtime_unavailable"
		return result
	}
	result.Enabled = result.ConfigExists
	if !result.Enabled {
		result.State, result.Error = CheckMissing, "ssh_not_configured"
		return result
	}
	payload, err := os.ReadFile(o.Paths.SSHPID())
	if err != nil {
		result.State, result.Error = CheckWarning, "ssh_pid_unavailable"
		return result
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(payload)))
	if err != nil || pid <= 0 {
		result.State, result.Error = CheckWarning, "ssh_pid_invalid"
		return result
	}
	result.PID = pid
	if !processIsMobdeskSSH(pid, result.ConfigPath) || !sshPortResponds(ctx, o.SSHPort) {
		result.State, result.Error = CheckWarning, "ssh_not_running"
		return result
	}
	result.Running, result.State = true, CheckOK
	return result
}

var ipv4Pattern = regexp.MustCompile(`^\s+inet\s+((?:[0-9]{1,3}\.){3}[0-9]{1,3})\b`)

func collectNetwork(ctx context.Context, o Options) NetworkStatus {
	result := NetworkStatus{State: CheckUnknown, Addresses: []string{}}
	if !commandAvailable(o, "ifconfig") {
		result.State, result.Error = CheckMissing, "ifconfig_missing"
		return result
	}
	command := runWithTimeout(ctx, o, "ifconfig")
	if command.Err != nil {
		result.Error = "ifconfig_failed"
		return result
	}
	interfaceName := ""
	for _, line := range strings.Split(string(command.Stdout), "\n") {
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				interfaceName = strings.TrimSuffix(fields[0], ":")
			}
		}
		match := ipv4Pattern.FindStringSubmatch(line)
		if len(match) != 2 || match[1] == "127.0.0.1" || net.ParseIP(match[1]) == nil {
			continue
		}
		if !contains(result.Addresses, match[1]) {
			if interfaceName == "wlan0" {
				result.Addresses = append([]string{match[1]}, result.Addresses...)
			} else {
				result.Addresses = append(result.Addresses, match[1])
			}
		}
	}
	if len(result.Addresses) == 0 {
		result.State, result.Error = CheckWarning, "no_ipv4_address"
		return result
	}
	result.State, result.Preferred = CheckOK, result.Addresses[0]
	return result
}

func collectTermuxAPIs(ctx context.Context, o Options) (BatteryStatus, WiFiStatus) {
	var battery BatteryStatus
	var wifi WiFiStatus
	var group sync.WaitGroup
	group.Add(2)
	go func() { defer group.Done(); battery = collectBattery(ctx, o) }()
	go func() { defer group.Done(); wifi = collectWiFi(ctx, o) }()
	group.Wait()
	return battery, wifi
}

func collectBattery(ctx context.Context, o Options) BatteryStatus {
	result := BatteryStatus{State: CheckMissing}
	if !commandAvailable(o, "termux-battery-status") {
		result.Error = "termux_api_missing"
		return result
	}
	command := runWithTimeoutFor(ctx, o, termuxTimeout, "termux-battery-status")
	if command.Err != nil {
		result.State, result.Error = CheckUnknown, "battery_api_failed"
		return result
	}
	var payload struct {
		Percentage  *int     `json:"percentage"`
		Status      string   `json:"status"`
		Plugged     string   `json:"plugged"`
		Temperature *float64 `json:"temperature"`
		Health      string   `json:"health"`
	}
	if json.Unmarshal(command.Stdout, &payload) != nil {
		result.State, result.Error = CheckUnknown, "battery_json_invalid"
		return result
	}
	result.State, result.Available, result.Percentage, result.Status, result.Plugged, result.Temperature, result.Health = CheckOK, true, payload.Percentage, payload.Status, payload.Plugged, payload.Temperature, payload.Health
	return result
}

func collectWiFi(ctx context.Context, o Options) WiFiStatus {
	result := WiFiStatus{State: CheckMissing}
	if !commandAvailable(o, "termux-wifi-connectioninfo") {
		result.Error = "termux_api_missing"
		return result
	}
	command := runWithTimeoutFor(ctx, o, termuxTimeout, "termux-wifi-connectioninfo")
	if command.Err != nil {
		result.State, result.Error = CheckUnknown, "wifi_api_failed"
		return result
	}
	var payload struct {
		SSID          string `json:"ssid"`
		IP            string `json:"ip"`
		LinkSpeedMbps *int   `json:"link_speed_mbps"`
		FrequencyMHz  *int   `json:"frequency_mhz"`
	}
	if json.Unmarshal(command.Stdout, &payload) != nil {
		result.State, result.Error = CheckUnknown, "wifi_json_invalid"
		return result
	}
	result.State, result.Available, result.SSID, result.IP, result.LinkSpeedMbps, result.FrequencyMHz = CheckOK, true, payload.SSID, payload.IP, payload.LinkSpeedMbps, payload.FrequencyMHz
	result.Connected = payload.IP != "" || payload.SSID != ""
	return result
}

func collectInstallations(o Options) []InstallationStatus {
	entries, err := os.ReadDir(o.Paths.InstallationsDir())
	if err != nil {
		return nil
	}
	values := make([]InstallationStatus, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		payload, err := os.ReadFile(filepath.Join(o.Paths.InstallationsDir(), entry.Name()))
		if err != nil {
			continue
		}
		var value InstallationStatus
		if json.Unmarshal(payload, &value) == nil && value.Name != "" {
			values = append(values, value)
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Name < values[j].Name })
	return values
}

func enrichInstallationMetadata(values []InstallationStatus) []InstallationStatus {
	profiles := install.Tools()
	for index := range values {
		if values[index].Source == "" {
			values[index].Source = "mobdesk"
		}
		values[index].Managed = values[index].Source == "mobdesk"
		for _, profile := range profiles {
			if values[index].Name == profile.Name {
				if values[index].StorageEstimate == nil {
					values[index].StorageEstimate = profile.StorageEstimate
				}
				break
			}
		}
	}
	return values
}

func reconcileInstallations(ctx context.Context, o Options, values []InstallationStatus) []InstallationStatus {
	values = enrichInstallationMetadata(values)
	installed := make(map[string]bool, len(values))
	for _, value := range values {
		if value.State == "installed" {
			installed[value.Name] = true
		}
	}
	for index := range values {
		profile, ok := install.Resolve(values[index].Name)
		if !ok || (values[index].State != "installed" && values[index].State != "partial") {
			continue
		}
		for _, dependency := range profile.Requires {
			if !installed[dependency] {
				values[index].MissingDependencies = append(values[index].MissingDependencies, dependency)
			}
		}
		for _, executable := range profileExecutables(profile) {
			if runWithTimeout(ctx, o, executable.Name, executable.VersionArg...).Err != nil {
				values[index].MissingExecutables = append(values[index].MissingExecutables, executable.Name)
			}
		}
		if len(values[index].MissingDependencies) > 0 || len(values[index].MissingExecutables) > 0 {
			values[index].State = "partial"
		}
	}
	return values
}

func profileExecutables(profile install.AppProfile) []install.ExecutableSpec {
	if len(profile.RequiredExecutables) > 0 {
		return profile.RequiredExecutables
	}
	if profile.Executable == "" {
		return nil
	}
	return []install.ExecutableSpec{{Name: profile.Executable, VersionArg: profile.VersionArg}}
}

func commandAvailable(o Options, name string) bool { _, err := o.LookPath(name); return err == nil }
func runWithTimeout(ctx context.Context, o Options, name string, args ...string) CommandResult {
	return runWithTimeoutFor(ctx, o, commandTimeout, name, args...)
}
func runWithTimeoutFor(ctx context.Context, o Options, timeout time.Duration, name string, args ...string) CommandResult {
	commandContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return o.CommandRunner.Run(commandContext, name, args...)
}
func fileExists(path string) bool { _, err := os.Stat(path); return err == nil }
func contains(values []string, value string) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}

func processIsMobdeskSSH(pid int, configPath string) bool {
	commandLine, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil || !strings.Contains(strings.ReplaceAll(string(commandLine), "\x00", " "), configPath) {
		return false
	}
	executable, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	return err == nil && filepath.Base(executable) == "sshd"
}

func sshPortResponds(ctx context.Context, port int) bool {
	dialer := net.Dialer{Timeout: 250 * time.Millisecond}
	connection, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return false
	}
	defer func() { _ = connection.Close() }()
	_ = connection.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buffer := make([]byte, 4)
	_, err = connection.Read(buffer)
	return err == nil && strings.HasPrefix(string(buffer), "SSH-")
}

func detectTermuxRuntime(prefix string) bool {
	if os.Getenv("TERMUX_VERSION") != "" {
		return true
	}
	return strings.HasPrefix(filepath.Clean(prefix), "/data/data/com.termux/files/usr")
}

func summarize(value SystemStatus) AlertSummary {
	states := []CheckState{value.Host.State, value.Setup.State, value.Workspace.State, value.Storage.State, value.SSH.State, value.Network.State, value.Battery.State, value.WiFi.State}
	var result AlertSummary
	for _, state := range states {
		switch state {
		case CheckOK:
			result.OK++
		case CheckWarning:
			result.Warnings++
		case CheckError:
			result.Errors++
		case CheckMissing:
			result.Missing++
		case CheckUnknown:
			result.Unknown++
		}
	}
	for _, installation := range value.Installations {
		if installation.State == "installed" {
			result.OK++
		} else if installation.State == "failed" || installation.State == "partial" {
			result.Warnings++
		} else {
			result.Unknown++
		}
	}
	return result
}

func overallState(value SystemStatus) OverallState {
	if value.Alerts.Errors > 0 {
		return StateError
	}
	if value.Alerts.Warnings > 0 {
		return StateDegraded
	}
	if value.Alerts.OK == 0 {
		return StateUnknown
	}
	return StateHealthy
}
