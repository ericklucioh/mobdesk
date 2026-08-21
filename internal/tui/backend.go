package tui

import (
	"context"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/ericklucioh/mobdesk/internal/status"
	"github.com/ericklucioh/mobdesk/internal/version"
)

// Backend is the communication boundary between the TUI and the Mobdesk
// services. The real backend delegates to the executable and collectors; the
// mock backend keeps the same messages while remaining safe to run anywhere.
type Backend interface {
	StatusCmd() tea.Cmd
	OperationCmd(args ...string) tea.Cmd
	ShellCmd() tea.Cmd
}

type realBackend struct {
	ctx    context.Context
	cancel context.CancelFunc
	locale i18n.Locale
}

func newRealBackend(locale i18n.Locale) *realBackend {
	ctx, cancel := context.WithCancel(context.Background())
	return &realBackend{ctx: ctx, cancel: cancel, locale: locale}
}

func (b *realBackend) StatusCmd() tea.Cmd {
	return runStatusCommand(b.ctx, b.locale)
}

func (b *realBackend) OperationCmd(args ...string) tea.Cmd {
	if len(args) > 0 && !containsArg(args, "--json") && (args[0] == "install" || args[0] == "setup") {
		return runInteractiveOperationWithLocale(b.ctx, b.locale, args...)
	}
	if len(args) > 0 && containsArg(args, "--progress") && (args[0] == "install" || args[0] == "uninstall") {
		return runInstallCommandWithLocale(b.ctx, b.locale, args...)
	}
	return runCommandWithLocale(b.ctx, b.locale, args...)
}

func (b *realBackend) ShellCmd() tea.Cmd {
	return realShellCommand(b.ctx, b.locale)
}

func (b *realBackend) Cancel() {
	b.cancel()
}

// NewMockBackend creates a production-build mock for manual visual testing.
// Supported scenarios are healthy, degraded and error. Unknown values use
// healthy so a typo never prevents the TUI from opening.
func NewMockBackend(scenario string) Backend {
	return newMockBackend(scenario, i18n.LocaleENUS)
}

func NewMockBackendLocale(scenario string, locale i18n.Locale) Backend {
	return newMockBackend(scenario, locale)
}

type mockBackend struct {
	mu        sync.Mutex
	scenario  string
	localizer i18n.Localizer
	status    status.SystemStatus
	info      version.Info
}

func newMockBackend(scenario string, locale i18n.Locale) *mockBackend {
	if scenario != "degraded" && scenario != "error" {
		scenario = "healthy"
	}
	value := mockStatus(scenario)
	return &mockBackend{scenario: scenario, localizer: i18n.New(locale), status: value, info: version.Current()}
}

func (m *mockBackend) StatusCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(120 * time.Millisecond)
		m.mu.Lock()
		defer m.mu.Unlock()
		value := m.status
		value.GeneratedAt = time.Now()
		return statusMessage{value: value, info: m.info}
	}
}

func (m *mockBackend) OperationCmd(args ...string) tea.Cmd {
	command := ""
	if len(args) > 0 {
		command = args[0]
	}
	checkOnly := containsArg(args, "--check")
	upgradeOnly := containsArg(args, "--upgrade-system")
	return func() tea.Msg {
		time.Sleep(700 * time.Millisecond)
		m.mu.Lock()
		defer m.mu.Unlock()
		if err := validateMockOperation(args, m.localizer); err != nil {
			return operationMessage{command: command, err: err}
		}

		if m.scenario == "error" {
			return operationMessage{
				command: command,
				err:     i18n.NewError(i18n.ErrorOperationFailed, "mock_operation_failed", map[string]any{"Detail": mockError(m.localizer, command)}, nil),
			}
		}

		result := operationResult{Success: true, State: "completed", Message: mockSuccess(m.localizer, command)}
		switch command {
		case "start":
			m.status.SSH.Running = true
			m.status.SSH.Enabled = true
			m.status.SSH.State = status.CheckOK
			m.status.Overall = status.StateHealthy
		case "stop":
			m.status.SSH.Running = false
			m.status.SSH.State = status.CheckWarning
			m.status.Overall = status.StateDegraded
		case "setup":
			m.status.Setup.Completed = true
			m.status.Setup.State = status.CheckOK
			m.status.Workspace.Exists = true
			m.status.Workspace.State = status.CheckOK
			if upgradeOnly {
				result.Message = m.localizer.Text(i18n.TUIOperationCompleted, nil)
			}
		case "update":
			result.CurrentVersion = m.info.Version
			result.LatestVersion = m.info.Version
			if checkOnly {
				result.State = "current"
				result.Message = m.localizer.Text(i18n.OutputUpdateCurrent, map[string]any{"Version": m.info.Version})
			} else {
				result.State = "updated"
				result.Updated = true
				result.LatestVersion = "mock-2.0"
				m.info.Version = result.LatestVersion
				result.Message = m.localizer.Text(i18n.TUIOperationCompleted, nil)
			}
		case "install":
			if len(args) > 1 {
				m.install(args[1])
				result.Language = args[1]
				result.Version = "mock-1.0"
			}
		case "uninstall":
			if len(args) > 1 {
				m.uninstall(args[1])
				result.Target = args[1]
			}
		}
		return operationMessage{command: command, result: result}
	}
}

func validateMockOperation(args []string, localizer i18n.Localizer) error {
	if len(args) == 0 {
		return i18n.NewError(i18n.ErrorInvalidArgs, "mock_invalid_args", map[string]any{"Detail": "missing command"}, nil)
	}
	command := args[0]
	switch command {
	case "start", "stop":
		return nil
	case "install":
		if len(args) < 2 || args[1] == "" {
			return i18n.NewError(i18n.ErrorInvalidArgs, "mock_invalid_args", map[string]any{"Detail": localizer.Text(i18n.CommandInstallUse, nil)}, nil)
		}
		return nil
	case "uninstall":
		if len(args) < 2 || args[1] == "" {
			return i18n.NewError(i18n.ErrorInvalidArgs, "mock_invalid_args", map[string]any{"Detail": localizer.Text(i18n.CommandUninstallUse, nil)}, nil)
		}
		return nil
	case "setup", "update":
		return nil
	default:
		return i18n.NewError(i18n.ErrorUnknownCommand, "mock_unknown_command", map[string]any{"Command": command}, nil)
	}
}

func containsArg(args []string, wanted string) bool {
	for _, arg := range args {
		if arg == wanted {
			return true
		}
	}
	return false
}

func (m *mockBackend) ShellCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(250 * time.Millisecond)
		return operationMessage{command: "shell", result: operationResult{
			Success: true,
			State:   "completed",
			Message: m.localizer.Text(i18n.TUIOperationCompleted, nil),
		}}
	}
}

func (m *mockBackend) install(name string) {
	for index := range m.status.Installations {
		if m.status.Installations[index].Name == name {
			m.status.Installations[index].State = "installed"
			m.status.Installations[index].Version = "mock-1.0"
			m.status.Installations[index].LastError = ""
			return
		}
	}
	for _, profile := range install.Catalog() {
		if profile.Name == name {
			m.status.Installations = append(m.status.Installations, status.InstallationStatus{
				Name:       profile.Name,
				Kind:       profile.Kind,
				Package:    profile.Package,
				Executable: profile.Executable,
				State:      "installed",
				Source:     "mobdesk",
				Managed:    true,
				Version:    "mock-1.0",
			})
			return
		}
	}
}

func (m *mockBackend) uninstall(name string) {
	for index := range m.status.Installations {
		if m.status.Installations[index].Name == name {
			m.status.Installations[index].State = "uninstalled"
			return
		}
	}
}

func mockStatus(scenario string) status.SystemStatus {
	value := status.SystemStatus{
		SchemaVersion: 2,
		Command:       "status",
		Success:       true,
		State:         string(status.StateHealthy),
		GeneratedAt:   time.Now(),
		Overall:       status.StateHealthy,
		Host:          status.HostStatus{State: status.CheckOK, Termux: true, OS: "Android", Architecture: "arm64", OpenSSH: true, Ifconfig: true, WakeLockAvailable: true},
		Setup:         status.SetupStatus{State: status.CheckOK, Completed: true, Phases: map[string]string{"directories": "done", "workspace-created": "done", "ssh-configured": "done"}},
		Workspace:     status.WorkspaceStatus{State: status.CheckOK, Exists: true, Path: "/data/data/com.termux/files/home/workspace"},
		Storage:       status.StorageStatus{State: status.CheckOK, DeviceTotal: 128 * 1024 * 1024 * 1024, DeviceFree: 42 * 1024 * 1024 * 1024},
		SSH:           status.SSHStatus{State: status.CheckOK, Enabled: true, Running: true, Port: 8022},
		Network:       status.NetworkStatus{State: status.CheckOK, Addresses: []string{"172.19.0.1", "172.18.0.1", "10.42.0.1"}, Preferred: "172.19.0.1"},
		Battery:       status.BatteryStatus{State: status.CheckOK, Available: true, Percentage: intPointer(72), Status: "normal", Plugged: "unplugged"},
		WiFi:          status.WiFiStatus{State: status.CheckOK, Available: true, Connected: true, SSID: "mobdesk-lab", IP: "172.19.0.1"},
		Java:          status.JavaStatus{State: status.CheckOK, Installed: true, Version: "openjdk 21", Home: "/data/data/com.termux/files/usr/lib/jvm/java-21-openjdk"},
		Alerts:        status.AlertSummary{OK: 12},
	}
	for _, profile := range install.Catalog() {
		value.Installations = append(value.Installations, status.InstallationStatus{
			Name: profile.Name, Kind: profile.Kind, Package: profile.Package, Executable: profile.Executable, State: "installed", Source: "mobdesk", Managed: true, Version: "mock-1.0",
		})
	}
	if scenario == "degraded" || scenario == "error" {
		value.Overall = status.StateDegraded
		value.State = string(status.StateDegraded)
		value.SSH.Running = false
		value.SSH.State = status.CheckWarning
		value.Battery.State = status.CheckWarning
		value.Battery.Percentage = intPointer(18)
		value.Battery.Status = "low"
		value.Workspace.Exists = false
		value.Workspace.State = status.CheckWarning
		value.Alerts = status.AlertSummary{OK: 7, Warnings: 5, Missing: 2}
		value.Installations = nil
	}
	if scenario == "error" {
		value.Overall = status.StateError
		value.State = string(status.StateError)
		value.Setup.State = status.CheckError
		value.Workspace.State = status.CheckError
		value.SSH.State = status.CheckError
		value.Network.State = status.CheckError
		value.Alerts = status.AlertSummary{OK: 3, Warnings: 4, Errors: 5}
	}
	return value
}

func intPointer(value int) *int { return &value }

func mockSuccess(localizer i18n.Localizer, command string) string {
	switch command {
	case "start":
		return localizer.Text(i18n.TUIOperationCompleted, nil)
	case "stop":
		return localizer.Text(i18n.TUIOperationCompleted, nil)
	case "setup":
		return localizer.Text(i18n.TUIOperationCompleted, nil)
	case "update":
		return localizer.Text(i18n.TUIOperationCompleted, nil)
	case "install":
		return localizer.Text(i18n.TUIOperationInstalled, map[string]any{"Name": "tool"})
	default:
		return localizer.Text(i18n.TUIOperationCompleted, nil)
	}
}

func mockError(localizer i18n.Localizer, command string) string {
	return localizer.Text(i18n.ErrorOperationFailed, map[string]any{"Detail": command})
}
