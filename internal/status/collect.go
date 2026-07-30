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
	SchemaVersion  = 1
	SSHPort        = 8022
	commandTimeout = 2 * time.Second
	termuxTimeout  = 1 * time.Second
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
	result := SystemStatus{
		SchemaVersion: SchemaVersion,
		GeneratedAt:   o.Now().UTC(),
		Host:          collectHost(o),
		Setup:         collectSetup(o),
		Storage:       collectStorage(ctx, o),
		Ubuntu:        collectUbuntu(ctx, o),
		SSH:           collectSSH(ctx, o),
		Network:       collectNetwork(ctx, o),
	}
	result.Battery, result.WiFi = collectTermuxAPIs(ctx, o)
	result.Installations = collectInstallations(o)
	result.Configurations = collectConfigurations(ctx, o)
	reconcileInstallationConfigurations(result.Installations, result.Configurations)
	result.Alerts = summarize(result)
	result.Overall = overallState(result)
	return result
}

// ReadInstallations returns the persisted installation records without
// running external commands. The TUI uses it as an immediate snapshot while
// the more expensive runtime status collection is still in progress.
func ReadInstallations(p paths.Paths) []InstallationStatus {
	o := Options{Paths: p}.withDefaults()
	installations := collectInstallations(o)
	reconcileInstallationConfigurations(installations, collectConfigurations(context.Background(), o))
	return installations
}

// IsTermuxRuntime reports whether the current process can control the Termux host.
// Commands started inside the Ubuntu PRoot must not invoke host operations.
func IsTermuxRuntime(prefix string) bool {
	return detectTermuxRuntime(prefix)
}

func collectHost(o Options) HostStatus {
	result := HostStatus{
		State:        CheckOK,
		Termux:       o.termux,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Home:         o.Paths.Home,
		Prefix:       o.Paths.Prefix,
	}
	result.ProotDistro = commandAvailable(o, "proot-distro")
	result.OpenSSH = commandAvailable(o, "sshd")
	result.Ifconfig = commandAvailable(o, "ifconfig")
	result.WakeLockAvailable = commandAvailable(o, "termux-wake-lock")
	result.TermuxAPIAvailable = commandAvailable(o, "termux-battery-status") || commandAvailable(o, "termux-wifi-connectioninfo")
	if !result.Termux {
		// Through Mobdesk SSH, this process runs inside Ubuntu/PRoot. It cannot
		// inspect the Termux control plane that launched it.
		result.State = CheckMissing
		result.Error = "termux_runtime_unavailable"
		return result
	}
	if _, err := os.Stat(o.Paths.Home); err != nil {
		result.State = CheckWarning
		result.Error = "home_unavailable"
	}
	return result
}

func collectSetup(o Options) SetupStatus {
	phases := []string{
		"directories", "packages-updated", "system-upgraded", "packages-installed",
		"ubuntu-installed", "workspace-created", "password-configured",
		"ssh-configured", "shell-configured", "launcher-installed",
	}
	requiredPhases := map[string]bool{
		"directories":         true,
		"packages-updated":    true,
		"packages-installed":  true,
		"ubuntu-installed":    true,
		"workspace-created":   true,
		"password-configured": true,
		"ssh-configured":      true,
		"shell-configured":    true,
		"launcher-installed":  true,
	}
	result := SetupStatus{State: CheckWarning, Phases: make(map[string]string, len(phases))}
	completed := true
	for _, phase := range phases {
		if _, err := os.Stat(o.Paths.SetupPhase(phase)); err == nil {
			result.Phases[phase] = "done"
			continue
		}
		result.Phases[phase] = "pending"
		if requiredPhases[phase] {
			completed = false
		}
	}
	result.Completed = completed
	if completed {
		result.State = CheckOK
	}
	return result
}

func collectStorage(ctx context.Context, o Options) StorageStatus {
	result := StorageStatus{State: CheckUnknown}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(o.Paths.Home, &stat); err != nil {
		result.Error = "device_storage_unavailable"
	} else {
		blockSize := int64(stat.Bsize)
		result.DeviceTotal = int64(stat.Blocks) * blockSize
		result.DeviceFree = int64(stat.Bavail) * blockSize
		result.DeviceUsed = result.DeviceTotal - int64(stat.Bfree)*blockSize
		result.State = CheckOK
	}
	if ctx.Err() != nil {
		result.State = CheckUnknown
		result.Error = "collection_cancelled"
		return result
	}
	return result
}

func collectUbuntu(ctx context.Context, o Options) UbuntuStatus {
	result := UbuntuStatus{State: CheckUnknown, WorkspacePath: o.Paths.UbuntuWorkspace()}
	if !o.termux {
		// The SSH session is already inside the Ubuntu workspace; proot-distro
		// exists only on the Termux side.
		result.Installed = true
		result.Accessible = true
		result.Workspace = directoryExists(result.WorkspacePath)
		if result.Workspace {
			result.State = CheckOK
		} else {
			result.State = CheckWarning
			result.Error = "workspace_missing"
		}
		return result
	}
	if !commandAvailable(o, "proot-distro") {
		result.State = CheckMissing
		result.Error = "proot_distro_missing"
		return result
	}
	result.Installed = commandSucceeds(ctx, o, "proot-distro", "login", "ubuntu", "--", "true")
	if !result.Installed {
		result.State = CheckError
		result.Error = "ubuntu_unavailable"
		return result
	}
	result.Accessible = true
	result.Workspace = commandSucceeds(ctx, o, "proot-distro", "login", "ubuntu", "--", "test", "-d", result.WorkspacePath)
	if result.Workspace {
		result.State = CheckOK
	} else {
		result.State = CheckWarning
		result.Error = "workspace_missing"
	}
	return result
}

func collectSSH(ctx context.Context, o Options) SSHStatus {
	configPath := o.Paths.SSHConfig()
	result := SSHStatus{
		State:        CheckUnknown,
		Port:         o.SSHPort,
		ConfigPath:   configPath,
		LogPath:      o.Paths.SSHLog(),
		ConfigExists: fileExists(configPath),
	}
	if !o.termux {
		result.State = CheckMissing
		result.Error = "termux_runtime_unavailable"
		return result
	}
	result.Enabled = result.ConfigExists
	if !result.Enabled {
		result.State = CheckMissing
		result.Error = "ssh_not_configured"
		return result
	}
	pidPath := o.Paths.SSHPID()
	pidBytes, err := os.ReadFile(pidPath)
	if err != nil {
		result.State = CheckWarning
		result.Error = "ssh_pid_unavailable"
		return result
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil || pid <= 0 {
		result.State = CheckWarning
		result.Error = "ssh_pid_invalid"
		return result
	}
	result.PID = pid
	if !processIsMobdeskSSH(pid, configPath) || !sshPortResponds(ctx, o.SSHPort) {
		result.State = CheckWarning
		result.Error = "ssh_not_running"
		return result
	}
	result.Running = true
	result.State = CheckOK
	return result
}

var ipv4Pattern = regexp.MustCompile(`^\s+inet\s+((?:[0-9]{1,3}\.){3}[0-9]{1,3})\b`)

func collectNetwork(ctx context.Context, o Options) NetworkStatus {
	result := NetworkStatus{State: CheckUnknown, Addresses: []string{}}
	if !commandAvailable(o, "ifconfig") {
		result.State = CheckMissing
		result.Error = "ifconfig_missing"
		return result
	}
	command := runWithTimeout(ctx, o, "ifconfig")
	if command.Err != nil {
		result.State = CheckUnknown
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
		result.State = CheckWarning
		result.Error = "no_ipv4_address"
		return result
	}
	result.State = CheckOK
	result.Preferred = result.Addresses[0]
	return result
}

func collectBattery(ctx context.Context, o Options) BatteryStatus {
	result := BatteryStatus{State: CheckMissing}
	if !commandAvailable(o, "termux-battery-status") {
		result.Error = "termux_api_missing"
		return result
	}
	command := runWithTimeoutFor(ctx, o, termuxTimeout, "termux-battery-status")
	if command.Err != nil {
		result.State = CheckUnknown
		result.Error = "battery_api_failed"
		return result
	}
	var payload struct {
		Percentage  *int     `json:"percentage"`
		Status      string   `json:"status"`
		Plugged     string   `json:"plugged"`
		Temperature *float64 `json:"temperature"`
		Health      string   `json:"health"`
	}
	if err := json.Unmarshal(command.Stdout, &payload); err != nil {
		result.State = CheckUnknown
		result.Error = "battery_json_invalid"
		return result
	}
	result.State, result.Available = CheckOK, true
	result.Percentage = payload.Percentage
	result.Status = payload.Status
	result.Plugged = payload.Plugged
	result.Temperature = payload.Temperature
	result.Health = payload.Health
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
		result.State = CheckUnknown
		result.Error = "wifi_api_failed"
		return result
	}
	var payload struct {
		SSID          string `json:"ssid"`
		IP            string `json:"ip"`
		LinkSpeedMbps *int   `json:"link_speed_mbps"`
		FrequencyMHz  *int   `json:"frequency_mhz"`
	}
	if err := json.Unmarshal(command.Stdout, &payload); err != nil {
		result.State = CheckUnknown
		result.Error = "wifi_json_invalid"
		return result
	}
	result.State, result.Available = CheckOK, true
	result.SSID = payload.SSID
	result.IP = payload.IP
	result.Connected = payload.IP != "" || payload.SSID != ""
	result.LinkSpeedMbps = payload.LinkSpeedMbps
	result.FrequencyMHz = payload.FrequencyMHz
	return result
}

func collectInstallations(o Options) []InstallationStatus {
	directory := o.Paths.InstallationsDir()
	entries, err := os.ReadDir(directory)
	if err != nil {
		return collectCatalogInstallations(o, nil)
	}
	result := make([]InstallationStatus, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		payload, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			continue
		}
		var installation InstallationStatus
		if err := json.Unmarshal(payload, &installation); err != nil || installation.Name == "" {
			continue
		}
		result = append(result, installation)
	}
	result = normalizeInstallationProvenance(result)
	result = enrichInstallationMetadata(result)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return collectCatalogInstallations(o, result)
}

func collectConfigurations(ctx context.Context, o Options) []ConfigurationStatus {
	profiles := install.DefaultConfigProfiles()
	byApp := make(map[string]ConfigurationStatus)
	for _, app := range install.Tools() {
		if app.ConfigProfile == "" {
			continue
		}
		profile, ok := profiles[app.ConfigProfile]
		if !ok {
			continue
		}
		byApp[app.Name] = ConfigurationStatus{
			App:          app.Name,
			Profile:      profile.ID,
			State:        ConfigStateNotApplied,
			ManagedPaths: append([]string(nil), profile.ManagedPaths...),
		}
	}

	entries, err := os.ReadDir(o.Paths.ConfigurationsDir())
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			payload, readErr := os.ReadFile(filepath.Join(o.Paths.ConfigurationsDir(), entry.Name()))
			if readErr != nil {
				continue
			}
			var record install.ConfigurationRecord
			if json.Unmarshal(payload, &record) != nil || record.App == "" {
				continue
			}
			state, modified := reconcileConfiguration(ctx, o, record)
			byApp[record.App] = ConfigurationStatus{
				App:           record.App,
				Profile:       record.Profile,
				State:         state,
				ManagedPaths:  append([]string(nil), record.ManagedPaths...),
				ModifiedPaths: modified,
				Conflicts:     append([]string(nil), record.Conflicts...),
			}
		}
	}
	for app, value := range byApp {
		if value.State != ConfigStateNotApplied {
			continue
		}
		for _, path := range value.ManagedPaths {
			if configurationPathExists(ctx, o, path) {
				value.State = ConfigStateConflict
				value.Conflicts = []string{path}
				byApp[app] = value
				break
			}
		}
	}

	result := make([]ConfigurationStatus, 0, len(byApp))
	for _, value := range byApp {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].App < result[j].App })
	return result
}

func reconciledConfigurationState(ctx context.Context, o Options, record install.ConfigurationRecord) ConfigState {
	state, _ := reconcileConfiguration(ctx, o, record)
	return state
}

func reconcileConfiguration(ctx context.Context, o Options, record install.ConfigurationRecord) (ConfigState, []string) {
	state := record.State
	if state == "" {
		state = ConfigStateNotApplied
	}
	if len(record.Conflicts) > 0 {
		return ConfigStateConflict, append([]string(nil), record.ModifiedPaths...)
	}
	if len(record.ModifiedPaths) > 0 {
		return ConfigStateModified, append([]string(nil), record.ModifiedPaths...)
	}
	if state != ConfigStateApplied {
		return state, nil
	}
	modified := configurationModifiedPaths(ctx, o, record)
	if len(modified) > 0 {
		return ConfigStateModified, modified
	}
	return state, nil
}

func configurationModifiedPaths(ctx context.Context, o Options, record install.ConfigurationRecord) []string {
	modified := make([]string, 0)
	for path, expected := range record.FileHashes {
		if !validConfigurationPath(path) || expected == "" {
			continue
		}
		current, ok := configurationHash(ctx, o, path)
		if ok && current != expected {
			modified = append(modified, path)
		}
	}
	return modified
}

func configurationPathExists(ctx context.Context, o Options, path string) bool {
	if !validConfigurationPath(path) {
		return false
	}
	if o.termux {
		return runWithTimeout(ctx, o, "proot-distro", "login", "ubuntu", "--", "test", "-e", path).Err == nil
	}
	return runWithTimeout(ctx, o, "test", "-e", path).Err == nil
}

func configurationHash(ctx context.Context, o Options, path string) (string, bool) {
	var result CommandResult
	if o.termux {
		result = runWithTimeout(ctx, o, "proot-distro", "login", "ubuntu", "--", "sha256sum", "--", path)
	} else {
		result = runWithTimeout(ctx, o, "sha256sum", "--", path)
	}
	if result.Err != nil {
		return "", false
	}
	fields := strings.Fields(string(result.Stdout))
	if len(fields) == 0 {
		return "", false
	}
	return fields[0], true
}

func validConfigurationPath(path string) bool {
	clean := filepath.Clean(path)
	return path != "" && clean == path && strings.HasPrefix(path, "/root/")
}

func reconcileInstallationConfigurations(installations []InstallationStatus, configurations []ConfigurationStatus) {
	states := make(map[string]ConfigState, len(configurations))
	for _, configuration := range configurations {
		states[configuration.App] = configuration.State
	}
	for index := range installations {
		if state, ok := states[installations[index].Name]; ok {
			installations[index].ConfigState = state
			continue
		}
		if profile, ok := install.Resolve(installations[index].Name); ok && profile.ConfigProfile != "" {
			installations[index].ConfigState = ConfigStateNotApplied
		} else {
			installations[index].ConfigState = ConfigStateUnavailable
		}
	}
}

func collectCatalogInstallations(o Options, persisted []InstallationStatus) []InstallationStatus {
	persisted = normalizeInstallationProvenance(persisted)
	persisted = enrichInstallationMetadata(persisted)
	if !o.termux || !commandAvailable(o, "proot-distro") {
		return persisted
	}
	tools := install.Tools()
	args := catalogStatusArgs(tools)
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	result := o.CommandRunner.Run(ctx, "proot-distro", args...)
	if result.Err != nil {
		return persisted
	}
	available := make(map[string]bool)
	for _, executable := range strings.Fields(string(result.Stdout)) {
		available[executable] = true
	}
	persistedNames := make(map[string]bool, len(persisted))
	for _, value := range persisted {
		persistedNames[value.Name] = true
	}
	for _, tool := range tools {
		if !available[tool.Executable] || persistedNames[tool.Name] {
			continue
		}
		persisted = append(persisted, InstallationStatus{
			Name:       tool.Name,
			Kind:       tool.Kind,
			Package:    tool.Package,
			Executable: tool.Executable,
			State:      "installed",
			Source:     "detected",
		})
	}
	persisted = normalizeInstallationProvenance(persisted)
	persisted = enrichInstallationMetadata(persisted)
	sort.Slice(persisted, func(i, j int) bool {
		return persisted[i].Name < persisted[j].Name
	})
	return persisted
}

func normalizeInstallationProvenance(values []InstallationStatus) []InstallationStatus {
	for index := range values {
		if values[index].Source == "" {
			values[index].Source = "mobdesk"
		}
		values[index].Managed = values[index].Source == "mobdesk"
	}
	return values
}

func enrichInstallationMetadata(values []InstallationStatus) []InstallationStatus {
	profiles := install.Tools()
	for index := range values {
		for _, profile := range profiles {
			matches := values[index].Name == profile.Name || values[index].Package == profile.Package || values[index].Executable == profile.Executable
			if !matches {
				continue
			}
			if values[index].StorageEstimate == nil {
				values[index].StorageEstimate = profile.StorageEstimate
			}
			break
		}
	}
	return values
}

func catalogStatusArgs(tools []install.AppProfile) []string {
	args := make([]string, 0, len(tools)+6)
	args = append(args, "login", "ubuntu", "--", "env", "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "sh", "-c", `PATH="$HOME/.local/bin:$PATH"; for tool do if command -v "$tool" >/dev/null 2>&1; then printf '%s\n' "$tool"; fi; done`, "mobdesk-status")
	for _, tool := range tools {
		args = append(args, tool.Executable)
	}
	return args
}

func collectTermuxAPIs(ctx context.Context, o Options) (BatteryStatus, WiFiStatus) {
	var (
		battery BatteryStatus
		wifi    WiFiStatus
		group   sync.WaitGroup
	)
	group.Add(2)
	go func() {
		defer group.Done()
		battery = collectBattery(ctx, o)
	}()
	go func() {
		defer group.Done()
		wifi = collectWiFi(ctx, o)
	}()
	group.Wait()
	return battery, wifi
}

func commandAvailable(o Options, name string) bool {
	_, err := o.LookPath(name)
	return err == nil
}

func commandSucceeds(ctx context.Context, o Options, name string, args ...string) bool {
	return runWithTimeout(ctx, o, name, args...).Err == nil
}

func runWithTimeout(ctx context.Context, o Options, name string, args ...string) CommandResult {
	return runWithTimeoutFor(ctx, o, commandTimeout, name, args...)
}

func runWithTimeoutFor(ctx context.Context, o Options, timeout time.Duration, name string, args ...string) CommandResult {
	commandContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return o.CommandRunner.Run(commandContext, name, args...)
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func detectTermuxRuntime(prefix string) bool {
	// The visible PRoot root is more reliable than inherited host variables.
	osRelease, err := os.ReadFile("/etc/os-release")
	if err == nil && strings.Contains(string(osRelease), "ID=ubuntu") {
		return false
	}
	if os.Getenv("TERMUX_VERSION") != "" {
		return true
	}
	if !strings.HasPrefix(filepath.Clean(prefix), "/data/data/com.termux/files/usr") {
		return false
	}
	return true
}

func contains(values []string, value string) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}

func summarize(status SystemStatus) AlertSummary {
	states := []CheckState{
		status.Host.State, status.Setup.State, status.Storage.State, status.Ubuntu.State,
		status.SSH.State, status.Network.State, status.Battery.State, status.WiFi.State,
	}
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
	for _, installation := range status.Installations {
		switch installation.State {
		case "installed":
			result.OK++
		case "failed", "partial":
			result.Warnings++
		default:
			result.Unknown++
		}
	}
	for _, configuration := range status.Configurations {
		switch configuration.State {
		case ConfigStateConflict, ConfigStateModified, ConfigStateFailed:
			result.Warnings++
		}
	}
	return result
}

func overallState(status SystemStatus) OverallState {
	if status.Alerts.Errors > 0 {
		return StateError
	}
	if status.Alerts.Warnings > 0 {
		return StateDegraded
	}
	if status.Alerts.OK == 0 {
		return StateUnknown
	}
	return StateHealthy
}
