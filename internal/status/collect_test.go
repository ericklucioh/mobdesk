package status

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/ericklucioh/mobdesk/internal/paths"
)

type fakeRunner struct {
	mu       sync.Mutex
	commands []string
	outputs  map[string]CommandResult
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) CommandResult {
	command := name + " " + strings.Join(args, " ")
	f.mu.Lock()
	f.commands = append(f.commands, command)
	f.mu.Unlock()
	if result, ok := f.outputs[command]; ok {
		return result
	}
	if name == "ifconfig" {
		return CommandResult{Stdout: []byte("wlan0: flags\n\tinet 192.168.1.20 netmask 255.255.255.0\n")}
	}
	return CommandResult{}
}

func availableCommands(names ...string) func(string) (string, error) {
	available := make(map[string]bool, len(names))
	for _, name := range names {
		available[name] = true
	}
	return func(name string) (string, error) {
		if available[name] {
			return "/usr/bin/" + name, nil
		}
		return "", errors.New("not found")
	}
}

func TestCollectProducesStableJSONAndUsesOptionalTermuxAPI(t *testing.T) {
	home := t.TempDir()
	prefix := t.TempDir()
	paths := paths.New(home, prefix)
	if err := os.MkdirAll(paths.StateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.SetupPhase("directories"), []byte("done"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{outputs: map[string]CommandResult{
		"termux-battery-status ":      {Stdout: []byte(`{"percentage":87,"status":"充电","plugged":"USB","temperature":31.5}`)},
		"termux-wifi-connectioninfo ": {Stdout: []byte(`{"ssid":"study","ip":"192.168.1.20","link_speed_mbps":433,"frequency_mhz":5180}`)},
	}}
	value := Collect(context.Background(), Options{
		Paths:         paths,
		CommandRunner: runner,
		LookPath:      availableCommands("proot-distro", "sshd", "ifconfig", "termux-wake-lock", "termux-battery-status", "termux-wifi-connectioninfo"),
		Now:           func() time.Time { return time.Date(2026, 7, 22, 15, 0, 0, 0, time.FixedZone("BRT", -3*60*60)) },
	})

	if value.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", value.SchemaVersion)
	}
	if value.Battery.State != CheckOK || value.Battery.Percentage == nil || *value.Battery.Percentage != 87 {
		t.Fatalf("unexpected battery status: %+v", value.Battery)
	}
	if value.WiFi.State != CheckOK || value.WiFi.IP != "192.168.1.20" {
		t.Fatalf("unexpected wifi status: %+v", value.WiFi)
	}
	if value.Storage.HomeBytes != nil || value.Storage.PrefixBytes != nil {
		t.Fatalf("quick status unexpectedly measured nested storage: %+v", value.Storage)
	}
	if value.Network.Preferred != "192.168.1.20" {
		t.Fatalf("preferred address = %q", value.Network.Preferred)
	}

	var document map[string]any
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatal(err)
	}
	if _, found := document["password"]; found {
		t.Fatal("status JSON exposed a password field")
	}
}

func TestCollectDegradesGracefullyWhenTermuxAPIIsMissing(t *testing.T) {
	runner := &fakeRunner{outputs: map[string]CommandResult{}}
	value := Collect(context.Background(), Options{
		Paths:         paths.New(t.TempDir(), t.TempDir()),
		CommandRunner: runner,
		LookPath:      availableCommands("ifconfig"),
	})

	if value.Battery.State != CheckMissing || value.Battery.Error != "termux_api_missing" {
		t.Fatalf("unexpected battery result: %+v", value.Battery)
	}
	if value.WiFi.State != CheckMissing || value.WiFi.Error != "termux_api_missing" {
		t.Fatalf("unexpected wifi result: %+v", value.WiFi)
	}
	if value.Overall == StateError {
		t.Fatalf("optional Termux API caused a global error: %+v", value)
	}
}

func TestSetupDoesNotRequireOptionalSystemUpgrade(t *testing.T) {
	home := t.TempDir()
	paths := paths.New(home, "")
	stateDir := paths.StateDir()
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{
		"directories", "packages-updated", "packages-installed", "ubuntu-installed",
		"workspace-created", "password-configured", "ssh-configured", "shell-configured", "launcher-installed",
	} {
		if err := os.WriteFile(paths.SetupPhase(phase), []byte("done"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	value := collectSetup(Options{Paths: paths}.withDefaults())
	if !value.Completed || value.State != CheckOK {
		t.Fatalf("optional system upgrade made setup incomplete: %+v", value)
	}
	if value.Phases["system-upgraded"] != "pending" {
		t.Fatalf("system-upgraded phase = %q, want pending", value.Phases["system-upgraded"])
	}
}

func TestCollectMarksInvalidTermuxJSONAsUnknown(t *testing.T) {
	runner := &fakeRunner{outputs: map[string]CommandResult{
		"termux-battery-status ":      {Stdout: []byte("not-json")},
		"termux-wifi-connectioninfo ": {Stdout: []byte("not-json")},
	}}
	value := Collect(context.Background(), Options{
		Paths:         paths.New(t.TempDir(), t.TempDir()),
		CommandRunner: runner,
		LookPath:      availableCommands("termux-battery-status", "termux-wifi-connectioninfo"),
	})

	if value.Battery.State != CheckUnknown || value.Battery.Error != "battery_json_invalid" {
		t.Fatalf("unexpected battery result: %+v", value.Battery)
	}
	if value.WiFi.State != CheckUnknown || value.WiFi.Error != "wifi_json_invalid" {
		t.Fatalf("unexpected wifi result: %+v", value.WiFi)
	}
}

func TestRenderTextAndJSON(t *testing.T) {
	value := SystemStatus{
		SchemaVersion: 1,
		GeneratedAt:   time.Date(2026, 7, 22, 18, 0, 0, 0, time.UTC),
		Overall:       StateHealthy,
		Host:          HostStatus{State: CheckOK, Architecture: "arm64"},
		Storage:       StorageStatus{State: CheckOK, DeviceTotal: 1024, DeviceFree: 512},
		Battery:       BatteryStatus{State: CheckMissing},
		WiFi:          WiFiStatus{State: CheckMissing},
	}
	var text bytes.Buffer
	if err := RenderText(&text, value); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text.String(), "Mobdesk status") || !strings.Contains(text.String(), "healthy") {
		t.Fatalf("unexpected human output: %s", text.String())
	}
	var document bytes.Buffer
	if err := EncodeJSON(&document, value); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(document.String(), "\n") {
		t.Fatal("JSON output does not end with a newline")
	}
	var decoded SystemStatus
	if err := json.Unmarshal(document.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
}

func TestCollectReadsGenericInstallationRecords(t *testing.T) {
	home := t.TempDir()
	directory := paths.New(home, "").InstallationsDir()
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	records := []InstallationStatus{
		{Name: "zeta-tool", Kind: "tool", State: "installed", Version: "1.0", LogPath: "/tmp/zeta.log"},
		{Name: "alpha-language", Kind: "language", State: "failed", LastError: "apt failed", LogPath: "/tmp/alpha.log"},
	}
	for _, record := range records {
		payload, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, record.Name+".json"), payload, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	value := Collect(context.Background(), Options{
		Paths:         paths.New(home, t.TempDir()),
		CommandRunner: &fakeRunner{outputs: map[string]CommandResult{}},
		LookPath:      availableCommands(),
	})
	if len(value.Installations) != 2 || value.Installations[0].Name != "alpha-language" {
		t.Fatalf("unexpected installations: %+v", value.Installations)
	}
	if value.Installations[0].LastError != "apt failed" || value.Installations[1].Kind != "tool" {
		t.Fatalf("installation metadata was not preserved: %+v", value.Installations)
	}
	if value.Installations[0].Source != "mobdesk" || !value.Installations[0].Managed {
		t.Fatalf("legacy installation provenance was not normalized: %+v", value.Installations[0])
	}
	if value.Overall != StateDegraded {
		t.Fatalf("overall = %s, want degraded", value.Overall)
	}
	var text bytes.Buffer
	if err := RenderText(&text, value); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text.String(), "alpha-language") || !strings.Contains(text.String(), "/tmp/alpha.log") {
		t.Fatalf("human output omitted installation diagnostics: %s", text.String())
	}
}

func TestCollectCatalogInstallationsRecognizesToolsWithoutRecords(t *testing.T) {
	tools := install.Tools()
	var executables []string
	for _, tool := range tools {
		if tool.Name == "git" || tool.Name == "cpp" {
			executables = append(executables, tool.Executable)
		}
	}
	output := strings.Join(executables, "\n") + "\n"
	command := "proot-distro login ubuntu -- env PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin sh -c PATH=\"$HOME/.local/bin:$PATH\"; for tool do if command -v \"$tool\" >/dev/null 2>&1; then printf '%s\\n' \"$tool\"; fi; done mobdesk-status " + strings.Join(executablesForTools(tools), " ")
	value := collectCatalogInstallations(Options{
		termux:   true,
		LookPath: availableCommands("proot-distro"),
		CommandRunner: &fakeRunner{outputs: map[string]CommandResult{
			command: {Stdout: []byte(output)},
		}},
	}, nil)

	for _, name := range []string{"git", "cpp"} {
		found := false
		for _, installation := range value {
			if installation.Name == name && installation.State == "installed" {
				found = true
			}
		}
		if !found {
			t.Fatalf("catalog did not recognize installed %s: %+v", name, value)
		}
	}
}

func TestCollectCatalogInstallationsAddsStorageEstimateToPersistedRecord(t *testing.T) {
	value := enrichInstallationMetadata([]InstallationStatus{{Name: "neovim", State: "installed"}})
	if len(value) != 1 || value[0].StorageEstimate == nil || value[0].StorageEstimate.TotalMinMB() != 15 {
		t.Fatalf("storage estimate was not added: %+v", value)
	}
}

func executablesForTools(tools []install.AppProfile) []string {
	executables := make([]string, 0, len(tools))
	for _, tool := range tools {
		executables = append(executables, tool.Executable)
	}
	return executables
}

func TestCatalogStatusUsesUbuntuOnlyPath(t *testing.T) {
	args := catalogStatusArgs(install.Tools())
	wantPrefix := []string{"login", "ubuntu", "--", "env", "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}
	if len(args) < len(wantPrefix) || strings.Join(args[:len(wantPrefix)], " ") != strings.Join(wantPrefix, " ") {
		t.Fatalf("catalog status args = %v", args)
	}
}
