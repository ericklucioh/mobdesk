package status

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	SchemaVersion  = 1
	SSHPort        = 8022
	commandTimeout = 2 * time.Second
	termuxTimeout  = 1 * time.Second
)

type Options struct {
	CommandRunner CommandRunner
	LookPath      func(string) (string, error)
	Now           func() time.Time
	Home          string
	Prefix        string
	SSHPort       int
}

func (o Options) withDefaults() Options {
	if o.CommandRunner == nil {
		o.CommandRunner = ExecRunner{}
	}
	if o.LookPath == nil {
		o.LookPath = exec.LookPath
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Home == "" {
		o.Home = os.Getenv("HOME")
	}
	if o.Home == "" {
		o.Home = "."
	}
	if o.Prefix == "" {
		o.Prefix = os.Getenv("PREFIX")
	}
	if o.Prefix == "" {
		o.Prefix = "/data/data/com.termux/files/usr"
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
	result.Alerts = summarize(result)
	result.Overall = overallState(result)
	return result
}

func collectHost(o Options) HostStatus {
	result := HostStatus{
		State:        CheckOK,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Home:         o.Home,
		Prefix:       o.Prefix,
	}
	result.ProotDistro = commandAvailable(o, "proot-distro")
	result.OpenSSH = commandAvailable(o, "sshd")
	result.Ifconfig = commandAvailable(o, "ifconfig")
	result.WakeLockAvailable = commandAvailable(o, "termux-wake-lock")
	result.TermuxAPIAvailable = commandAvailable(o, "termux-battery-status") || commandAvailable(o, "termux-wifi-connectioninfo")
	if _, err := os.Stat(o.Home); err != nil {
		result.State = CheckWarning
		result.Error = "home_unavailable"
	}
	return result
}

func collectSetup(o Options) SetupStatus {
	phases := []string{
		"directories", "packages-updated", "system-upgraded", "packages-installed",
		"ubuntu-installed", "workspace-created", "password-configured",
		"ssh-configured", "launcher-installed",
	}
	requiredPhases := map[string]bool{
		"directories":         true,
		"packages-updated":    true,
		"packages-installed":  true,
		"ubuntu-installed":    true,
		"workspace-created":   true,
		"password-configured": true,
		"ssh-configured":      true,
		"launcher-installed":  true,
	}
	result := SetupStatus{State: CheckWarning, Phases: make(map[string]string, len(phases))}
	completed := true
	for _, phase := range phases {
		if _, err := os.Stat(filepath.Join(o.Home, ".local", "share", "mobdesk", "state", phase+".done")); err == nil {
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
	if err := syscall.Statfs(o.Home, &stat); err != nil {
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
	result := UbuntuStatus{State: CheckUnknown, WorkspacePath: "/root/workspace"}
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
	home := o.Home
	runtimeDir := filepath.Join(home, ".local", "share", "mobdesk", "ssh")
	configPath := filepath.Join(home, ".config", "mobdesk", "ssh", "sshd_config")
	result := SSHStatus{
		State:        CheckUnknown,
		Port:         o.SSHPort,
		ConfigPath:   configPath,
		LogPath:      filepath.Join(runtimeDir, "sshd.log"),
		ConfigExists: fileExists(configPath),
	}
	result.Enabled = result.ConfigExists
	if !result.Enabled {
		result.State = CheckMissing
		result.Error = "ssh_not_configured"
		return result
	}
	pidPath := filepath.Join(runtimeDir, "sshd.pid")
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
	directory := filepath.Join(o.Home, ".local", "share", "mobdesk", "state", "installations")
	entries, err := os.ReadDir(directory)
	if err != nil {
		return []InstallationStatus{}
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
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
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
	defer connection.Close()
	_ = connection.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buffer := make([]byte, 4)
	_, err = connection.Read(buffer)
	return err == nil && strings.HasPrefix(string(buffer), "SSH-")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
