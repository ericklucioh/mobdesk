package tui

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
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
}

func newRealBackend() *realBackend {
	ctx, cancel := context.WithCancel(context.Background())
	return &realBackend{ctx: ctx, cancel: cancel}
}

func (b *realBackend) StatusCmd() tea.Cmd {
	return runStatusCommand(b.ctx)
}

func (b *realBackend) OperationCmd(args ...string) tea.Cmd {
	if len(args) > 0 && containsArg(args, "--progress") && (args[0] == "install" || args[0] == "uninstall" || args[0] == "config") {
		return runInstallCommand(b.ctx, args...)
	}
	return runCommand(b.ctx, args...)
}

func (b *realBackend) ShellCmd() tea.Cmd {
	return realShellCommand(b.ctx)
}

func (b *realBackend) Cancel() {
	b.cancel()
}

// NewMockBackend creates a production-build mock for manual visual testing.
// Supported scenarios are healthy, degraded and error. Unknown values use
// healthy so a typo never prevents the TUI from opening.
func NewMockBackend(scenario string) Backend {
	return newMockBackend(scenario)
}

type mockBackend struct {
	mu       sync.Mutex
	scenario string
	status   status.SystemStatus
	info     version.Info
}

func newMockBackend(scenario string) *mockBackend {
	if scenario != "degraded" && scenario != "error" {
		scenario = "healthy"
	}
	value := mockStatus(scenario)
	return &mockBackend{scenario: scenario, status: value, info: version.Current()}
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
		if err := validateMockOperation(args); err != nil {
			return operationMessage{command: command, err: err}
		}

		if m.scenario == "error" {
			return operationMessage{
				command: command,
				err:     errors.New(mockError(command)),
			}
		}

		result := operationResult{Success: true, State: "completed", Message: mockSuccess(command)}
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
			m.status.Ubuntu.Installed = true
			m.status.Ubuntu.Accessible = true
			m.status.Ubuntu.Workspace = true
			if upgradeOnly {
				result.Message = "Upgrade mock concluído"
			}
		case "update":
			result.CurrentVersion = m.info.Version
			result.LatestVersion = m.info.Version
			if checkOnly {
				result.State = "current"
				result.Message = "Nenhuma atualização pendente"
			} else {
				result.State = "updated"
				result.Updated = true
				result.LatestVersion = "mock-2.0"
				m.info.Version = result.LatestVersion
				result.Message = "Atualização mock concluída"
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
		case "config":
			if len(args) > 2 {
				m.configure(args[1], args[2])
				result.Target = args[2]
				result.Action = args[1]
			}
		}
		return operationMessage{command: command, result: result}
	}
}

func validateMockOperation(args []string) error {
	if len(args) == 0 {
		return errors.New("operação mock sem comando")
	}
	command := args[0]
	switch command {
	case "start", "stop":
		return nil
	case "install":
		if len(args) < 2 || args[1] == "" {
			return errors.New("mobdesk install exige uma linguagem")
		}
		return nil
	case "uninstall":
		if len(args) < 2 || args[1] == "" {
			return errors.New("mobdesk uninstall exige um app")
		}
		return nil
	case "config":
		if len(args) < 3 || (args[1] != "apply" && args[1] != "remove") || args[2] == "" {
			return errors.New("mobdesk config exige apply/remove e um app")
		}
		return nil
	case "setup", "update":
		return nil
	default:
		return fmt.Errorf("comando mock inexistente no CLI: %s", command)
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
			Message: "Shell mock encerrado",
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
	for _, language := range install.Languages() {
		if language.Name == name {
			m.status.Installations = append(m.status.Installations, status.InstallationStatus{
				Name:       language.Name,
				Kind:       "language",
				Package:    language.Package,
				Executable: language.Executable,
				State:      "installed",
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

func (m *mockBackend) configure(action, name string) {
	for index := range m.status.Configurations {
		if m.status.Configurations[index].App == name {
			if action == "apply" {
				m.status.Configurations[index].State = status.ConfigStateApplied
			} else {
				m.status.Configurations[index].State = status.ConfigStateRemoved
			}
			return
		}
	}
	state := status.ConfigStateApplied
	if action == "remove" {
		state = status.ConfigStateRemoved
	}
	m.status.Configurations = append(m.status.Configurations, status.ConfigurationStatus{App: name, Profile: "lazyvim", State: state})
}

func mockStatus(scenario string) status.SystemStatus {
	value := status.SystemStatus{
		SchemaVersion: 1,
		GeneratedAt:   time.Now(),
		Overall:       status.StateHealthy,
		Host:          status.HostStatus{State: status.CheckOK, Termux: true, OS: "Android", Architecture: "arm64", ProotDistro: true, OpenSSH: true, Ifconfig: true, WakeLockAvailable: true},
		Setup:         status.SetupStatus{State: status.CheckOK, Completed: true, Phases: map[string]string{"directories": "completed", "ubuntu-installed": "completed", "workspace-created": "completed", "ssh-configured": "completed"}},
		Storage:       status.StorageStatus{State: status.CheckOK, DeviceTotal: 128 * 1024 * 1024 * 1024, DeviceFree: 42 * 1024 * 1024 * 1024},
		Ubuntu:        status.UbuntuStatus{State: status.CheckOK, Installed: true, Accessible: true, Workspace: true, WorkspacePath: "/root/workspace"},
		SSH:           status.SSHStatus{State: status.CheckOK, Enabled: true, Running: true, Port: 8022},
		Network:       status.NetworkStatus{State: status.CheckOK, Addresses: []string{"172.19.0.1", "172.18.0.1", "10.42.0.1"}, Preferred: "172.19.0.1"},
		Battery:       status.BatteryStatus{State: status.CheckOK, Available: true, Percentage: intPointer(72), Status: "normal", Plugged: "unplugged"},
		WiFi:          status.WiFiStatus{State: status.CheckOK, Available: true, Connected: true, SSID: "mobdesk-lab", IP: "172.19.0.1"},
		Alerts:        status.AlertSummary{OK: 12},
	}
	for _, language := range install.Languages() {
		value.Installations = append(value.Installations, status.InstallationStatus{
			Name: language.Name, Kind: "language", Package: language.Package, Executable: language.Executable, State: "installed", Version: "mock-1.0",
		})
	}
	if scenario == "degraded" || scenario == "error" {
		value.Overall = status.StateDegraded
		value.SSH.Running = false
		value.SSH.State = status.CheckWarning
		value.Battery.State = status.CheckWarning
		value.Battery.Percentage = intPointer(18)
		value.Battery.Status = "low"
		value.Ubuntu.Workspace = false
		value.Alerts = status.AlertSummary{OK: 7, Warnings: 5, Missing: 2}
		value.Installations = nil
	}
	if scenario == "error" {
		value.Overall = status.StateError
		value.Setup.State = status.CheckError
		value.Ubuntu.State = status.CheckError
		value.SSH.State = status.CheckError
		value.Network.State = status.CheckError
		value.Alerts = status.AlertSummary{OK: 3, Warnings: 4, Errors: 5}
	}
	return value
}

func intPointer(value int) *int { return &value }

func mockSuccess(command string) string {
	switch command {
	case "start":
		return "Workstation mock iniciada"
	case "stop":
		return "Workstation mock parada"
	case "setup":
		return "Setup mock concluído"
	case "update":
		return "Atualização mock concluída"
	case "install":
		return "Ferramenta mock instalada"
	default:
		return fmt.Sprintf("Operação mock %s concluída", command)
	}
}

func mockError(command string) string {
	return fmt.Sprintf("falha simulada na operação %s", command)
}
