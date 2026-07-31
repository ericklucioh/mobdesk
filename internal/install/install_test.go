package install

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/paths"
)

type fakeRunner struct {
	commands []string
	results  map[string][]CommandResult
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) CommandResult {
	command := name + " " + strings.Join(args, " ")
	f.commands = append(f.commands, command)
	if results, ok := f.results[command]; ok && len(results) > 0 {
		result := results[0]
		f.results[command] = results[1:]
		return result
	}
	return CommandResult{Err: errors.New("command not configured")}
}

func testOptions(t *testing.T, runner CommandRunner) Options {
	t.Helper()
	base := t.TempDir()
	return Options{
		Paths:  paths.New(base, ""),
		Runner: runner,
		Now:    func() time.Time { return time.Date(2026, 7, 22, 18, 0, 0, 0, time.UTC) },
	}
}

func TestAppProfileContract(t *testing.T) {
	measuredAt := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	profile := AppProfile{
		Name:        "neovim",
		Aliases:     []string{"nvim"},
		Description: "editor modal",
		Package:     "neovim",
		Executable:  "nvim",
		VersionArg:  []string{"--version"},
		InstallKind: "apt",
		StorageEstimate: &StorageEstimate{
			AppMinMB:     20,
			AppMaxMB:     40,
			Source:       "fixture",
			Version:      "0.11",
			Architecture: "arm64",
			MeasuredAt:   measuredAt,
		},
	}

	if profile.Name != "neovim" || profile.StorageEstimate == nil || profile.StorageEstimate.MeasuredAt != measuredAt {
		t.Fatalf("unexpected app profile: %+v", profile)
	}
}

func TestCanonicalAppAndConfigStates(t *testing.T) {
	appStates := []AppState{
		AppStateAvailable,
		AppStateInstalling,
		AppStateInstalled,
		AppStateUninstalling,
		AppStateUninstalled,
		AppStatePartial,
		AppStateFailed,
	}
	configStates := []ConfigState{
		ConfigStateUnavailable,
		ConfigStateNotApplied,
		ConfigStateApplying,
		ConfigStateApplied,
		ConfigStateRemoving,
		ConfigStateRemoved,
		ConfigStateModified,
		ConfigStateConflict,
		ConfigStateFailed,
	}
	if len(appStates) != 7 || len(configStates) != 9 {
		t.Fatalf("canonical state set changed: app=%d config=%d", len(appStates), len(configStates))
	}
}

func TestCatalogProfilesDeclareDescriptionAndStorageEstimate(t *testing.T) {
	profiles := Tools()
	if len(profiles) != 26 {
		t.Fatalf("catalog has %d profiles, want 26", len(profiles))
	}
	seen := make(map[string]bool, len(profiles))
	for _, profile := range profiles {
		if profile.Name == "" || seen[profile.Name] {
			t.Fatalf("catalog has invalid or duplicate profile: %+v", profile)
		}
		seen[profile.Name] = true
		if profile.Description == "" || profile.InstallKind == "" || profile.StorageEstimate == nil {
			t.Fatalf("profile lacks required catalog metadata: %+v", profile)
		}
		estimate := profile.StorageEstimate
		if estimate.AppMinMB > estimate.AppMaxMB || estimate.DependenciesMinMB > estimate.DependenciesMaxMB || estimate.ConfigMinMB > estimate.ConfigMaxMB {
			t.Fatalf("invalid storage interval for %s: %+v", profile.Name, estimate)
		}
		if estimate.Source != "planning" || estimate.Architecture != "arm64" {
			t.Fatalf("storage estimate metadata missing for %s: %+v", profile.Name, estimate)
		}
		for _, alias := range profile.Aliases {
			if alias != strings.ToLower(alias) {
				t.Fatalf("alias %q for %s is not normalized", alias, profile.Name)
			}
		}
	}
}

func TestCatalogProfilesUseSelectedLocale(t *testing.T) {
	english := Tools(i18n.New(i18n.LocaleENUS))[0].Description
	brazilianPortuguese := Tools(i18n.New(i18n.LocalePTBR))[0].Description
	if english != i18n.New(i18n.LocaleENUS).Text(i18n.AppGoDescription, nil) || brazilianPortuguese != i18n.New(i18n.LocalePTBR).Text(i18n.AppGoDescription, nil) || english == brazilianPortuguese {
		t.Fatalf("localized descriptions = %q / %q", english, brazilianPortuguese)
	}
	profile := DefaultConfigProfiles(i18n.New(i18n.LocalePTBR))["lazyvim"]
	if profile.Description != i18n.New(i18n.LocalePTBR).Text(i18n.ProfileLazyVimDescription, nil) {
		t.Fatalf("localized config profile description = %q", profile.Description)
	}
}

func TestLegacyInstallationRecordRemainsReadable(t *testing.T) {
	options := testOptions(t, &fakeRunner{})
	if err := os.MkdirAll(options.Paths.InstallationsDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	payload := `{"name":"go","kind":"language","package":"golang","state":"failed","last_error":"historical failure","log_path":"/safe/go.log"}`
	if err := os.WriteFile(filepath.Join(options.Paths.InstallationsDir(), "go.json"), []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	record, err := loadInstallationRecord(options.Paths, "go")
	if err != nil || record.LastError != "historical failure" || record.LastErrorCode != "" {
		t.Fatalf("legacy record = %+v, error = %v", record, err)
	}
}

func TestStorageEstimateTotals(t *testing.T) {
	estimate := StorageEstimate{AppMinMB: 15, AppMaxMB: 30, DependenciesMinMB: 2, DependenciesMaxMB: 20, ConfigMinMB: 1, ConfigMaxMB: 5}
	if estimate.TotalMinMB() != 18 || estimate.TotalMaxMB() != 55 {
		t.Fatalf("totals = %d-%d, want 18-55", estimate.TotalMinMB(), estimate.TotalMaxMB())
	}
}

func TestResolveLanguagesAndAliases(t *testing.T) {
	for _, name := range []string{"go", "golang", "python", "python3", "node", "nodejs", "c", "c-lang", "cpp", "c++", "cplusplus", "lua", "lua5.4"} {
		if _, ok := Resolve(name); !ok {
			t.Fatalf("Resolve(%q) returned false", name)
		}
	}
	if _, ok := Resolve("rust"); ok {
		t.Fatal("Resolve(rust) unexpectedly succeeded")
	}
}

func TestResolveToolsAndPrerequisites(t *testing.T) {
	for _, name := range []string{"git", "gh", "tmux", "zellij", "micro", "lazygit", "tree", "ttt", "posting", "yazi", "tuifi", "htop", "ncdu", "inxi", "speedtest-cli", "opencode-cli", "codex-cli", "claudecode-cli", "leetgo"} {
		tool, ok := Resolve(name)
		if !ok || tool.Kind == "" || tool.InstallKind == "" {
			t.Fatalf("Resolve(%q) = %+v, %t", name, tool, ok)
		}
	}
	for _, unsupported := range []string{"clin", "glint"} {
		if _, ok := Resolve(unsupported); ok {
			t.Fatalf("Resolve(%q) unexpectedly succeeded", unsupported)
		}
	}
}

func TestFileManagersUsePinnedInstallProfiles(t *testing.T) {
	yazi, ok := Resolve("yazi")
	if !ok || yazi.InstallKind != "script" || !yazi.UserBin || !strings.Contains(yazi.Script, "v26.5.6") || !strings.Contains(yazi.Script, "sha256sum -c") {
		t.Fatalf("unexpected yazi profile: %+v", yazi)
	}
	if yazi.Executable != "yazi" || !slices.Contains(yazi.VersionArg, "--version") {
		t.Fatalf("unexpected yazi verification: %+v", yazi)
	}

	tuifi, ok := Resolve("tuifimanager")
	if !ok || tuifi.InstallKind != "script" || tuifi.UserBin || tuifi.Package != "TUIFIManager==5.2.6" || !slices.Contains(tuifi.Requires, "python") {
		t.Fatalf("unexpected tuifi profile: %+v", tuifi)
	}
	if !strings.Contains(tuifi.Script, "pipx install --force TUIFIManager==5.2.6") || tuifi.Executable != "tuifi" {
		t.Fatalf("unexpected tuifi installation: %+v", tuifi)
	}
}

func TestNeovimProfileUsesOptionalLazyVimConfiguration(t *testing.T) {
	neovim, ok := Resolve("neovim")
	if !ok {
		t.Fatal("neovim missing from catalog")
	}
	alias, ok := Resolve("nvim")
	if !ok || alias.Name != neovim.Name {
		t.Fatalf("nvim alias did not resolve to neovim: %+v", alias)
	}
	if neovim.Package != "neovim" || neovim.Executable != "nvim" || !slices.Equal(neovim.VersionArg, []string{"--version"}) || neovim.InstallKind != "apt" {
		t.Fatalf("unexpected neovim installation profile: %+v", neovim)
	}
	if neovim.ConfigProfile != "lazyvim" || neovim.ConfigTarget != "/root/.config/nvim" || neovim.MinimumVersion == "" || neovim.ProfileVersion == "" {
		t.Fatalf("unexpected neovim configuration profile: %+v", neovim)
	}
}

func TestInstallNeovimUsesUbuntuApt(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " apt-get -o DPkg::Lock::Timeout=300 install -y neovim": {{}},
	}}
	neovim, ok := Resolve("neovim")
	if !ok {
		t.Fatal("neovim missing from catalog")
	}
	if result := installTool(context.Background(), runner, time.Minute, t.TempDir()+"/install.log", neovim); result.Err != nil {
		t.Fatal(result.Err)
	}
	if !slices.Equal(runner.commands, []string{"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " apt-get -o DPkg::Lock::Timeout=300 install -y neovim"}) {
		t.Fatalf("commands = %v", runner.commands)
	}
}

func TestInstallToolUsesPrivateNPMPrefix(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " apt-get -o DPkg::Lock::Timeout=300 install -y npm":                   {{}},
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " env NPM_CONFIG_PREFIX=/root/.local npm install --yes -g opencode-ai": {{}},
	}}
	tool, ok := Resolve("opencode-cli")
	if !ok {
		t.Fatal("opencode-cli missing from catalog")
	}
	if result := installTool(context.Background(), runner, time.Minute, t.TempDir()+"/install.log", tool); result.Err != nil {
		t.Fatal(result.Err)
	}
	if len(runner.commands) != 2 {
		t.Fatalf("commands = %v", runner.commands)
	}
}

func TestInstallReleaseToolsUsePinnedArchitectureAwareBinaries(t *testing.T) {
	for _, name := range []string{"lazygit", "leetgo"} {
		t.Run(name, func(t *testing.T) {
			tool, ok := Resolve(name)
			if !ok {
				t.Fatalf("%s missing from catalog", name)
			}
			if tool.InstallKind != "script" || tool.UserBin || len(tool.Requires) != 0 {
				t.Fatalf("%s has unexpected installation profile: %+v", name, tool)
			}
			if !strings.Contains(tool.Script, "uname -m") || !strings.Contains(tool.Script, "sha256sum -c") || !strings.Contains(tool.Script, "github.com/") {
				t.Fatalf("%s script is not an architecture-aware verified release install: %s", name, tool.Script)
			}

			command := "proot-distro login ubuntu -- env PATH=" + ubuntuPath + " sh -ec " + tool.Script
			runner := &fakeRunner{results: map[string][]CommandResult{command: {{}}}}
			if result := installTool(context.Background(), runner, time.Minute, t.TempDir()+"/install.log", tool); result.Err != nil {
				t.Fatal(result.Err)
			}
			if !slices.Equal(runner.commands, []string{command}) {
				t.Fatalf("commands = %v, want %v", runner.commands, []string{command})
			}
		})
	}
}

func TestInstallNodeInstallsNPM(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " apt-get -o DPkg::Lock::Timeout=300 install -y nodejs npm": {{}},
	}}
	tool, ok := Resolve("node")
	if !ok {
		t.Fatal("node missing from catalog")
	}
	if result := installTool(context.Background(), runner, time.Minute, t.TempDir()+"/install.log", tool); result.Err != nil {
		t.Fatal(result.Err)
	}
}

func TestInstallTTTInstallsRequiredTools(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " apt-get -o DPkg::Lock::Timeout=300 install -y git ripgrep":                     {{}},
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " env GOBIN=/usr/local/bin go install github.com/eugenioenko/ttt/cmd/ttt@v1.1.0": {{}},
	}}
	tool, ok := Resolve("ttt")
	if !ok {
		t.Fatal("ttt missing from catalog")
	}
	if result := installTool(context.Background(), runner, time.Minute, t.TempDir()+"/install.log", tool); result.Err != nil {
		t.Fatal(result.Err)
	}
	if len(runner.commands) != 2 {
		t.Fatalf("commands = %v", runner.commands)
	}
}

func TestRunToolVersionAddsUserBinToPath(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " sh -ec PATH=\"$HOME/.local/bin:$PATH\"; exec \"$@\" -- zellij --version": {{Stdout: []byte("zellij 0.44.3")}},
	}}
	tool, ok := Resolve("zellij")
	if !ok {
		t.Fatal("zellij missing from catalog")
	}
	result := runToolVersion(context.Background(), runner, time.Minute, t.TempDir()+"/install.log", tool)
	if result.Err != nil || string(result.Stdout) != "zellij 0.44.3" {
		t.Fatalf("result = %+v, commands = %v", result, runner.commands)
	}
}

func TestRunUbuntuLoggedRestrictsPathToUbuntu(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " go version": {{Stdout: []byte("go version go1.26.5 linux/arm64\n")}},
	}}
	result := runUbuntuLogged(context.Background(), runner, time.Minute, t.TempDir()+"/install.log", "go", "version")
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if len(runner.commands) != 1 || runner.commands[0] != "proot-distro login ubuntu -- env PATH="+ubuntuPath+" go version" {
		t.Fatalf("commands = %v", runner.commands)
	}
}

func TestResolveNativeLanguageProfiles(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		executable  string
		versionArg  string
	}{
		{name: "c", packageName: "clang", executable: "clang", versionArg: "--version"},
		{name: "cpp", packageName: "clang", executable: "clang++", versionArg: "--version"},
		{name: "lua", packageName: "lua5.4", executable: "lua5.4", versionArg: "-v"},
	}
	for _, test := range tests {
		language, ok := Resolve(test.name)
		if !ok {
			t.Fatalf("Resolve(%q) returned false", test.name)
		}
		if language.Package != test.packageName || language.Executable != test.executable || language.VersionArg[0] != test.versionArg {
			t.Fatalf("Resolve(%q) = %+v", test.name, language)
		}
	}
}

func TestInstallSkipsAlreadyInstalledLanguage(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " python3 --version": {{Stdout: []byte("Python 3.12.1\n")}},
	}}
	result, err := Install(context.Background(), "python", testOptions(t, runner))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Installed || result.Changed || result.Version != "Python 3.12.1" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(runner.commands) != 1 {
		t.Fatalf("commands = %v, want one version check", runner.commands)
	}
}

func TestInstallUpdatesAndInstallsMissingLanguage(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " go version":                                           {{Err: errors.New("not found")}, {Stdout: []byte("go version go1.26.5 linux/arm64\n")}},
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " apt-get -o DPkg::Lock::Timeout=300 update":            {{}},
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " apt-get -o DPkg::Lock::Timeout=300 install -y golang": {{}},
	}}
	options := testOptions(t, runner)
	var updates []string
	options.Progress = func(message string) { updates = append(updates, message) }
	result, err := Install(context.Background(), "golang", options)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Installed || !result.Changed || result.Language != "go" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(runner.commands) != 4 {
		t.Fatalf("commands = %v, want version, update, install, version", runner.commands)
	}
	wantUpdates := []string{"verify go", "update Ubuntu indexes for go", "wait for the package manager", "install go", "verify go"}
	if !slices.Equal(updates, wantUpdates) {
		t.Fatalf("progress = %v, want %v", updates, wantUpdates)
	}
}

func TestInstallPersistsRecordAndCommandLog(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " node --version": {{Stdout: []byte("v22.1.0\n")}},
	}}
	options := testOptions(t, runner)
	result, err := Install(context.Background(), "node", options)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != "installed" || result.LogPath == "" {
		t.Fatalf("unexpected result: %+v", result)
	}
	recordBytes, err := os.ReadFile(filepath.Join(options.Paths.InstallationsDir(), "node.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(recordBytes), `"state": "installed"`) {
		t.Fatalf("record did not contain installed state: %s", recordBytes)
	}
	var record InstallationRecord
	if err := json.Unmarshal(recordBytes, &record); err != nil {
		t.Fatal(err)
	}
	if record.Source != "mobdesk" || record.Strategy != "node" || !slices.Contains(record.InstalledPackages, "nodejs") {
		t.Fatalf("installation provenance was not persisted: %+v", record)
	}
	logBytes, err := os.ReadFile(result.LogPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logBytes), "node --version") || !strings.Contains(string(logBytes), "v22.1.0") {
		t.Fatalf("log did not contain command output: %s", logBytes)
	}
}

type contextRunner struct{}

func (contextRunner) Run(ctx context.Context, _ string, _ ...string) CommandResult {
	<-ctx.Done()
	return CommandResult{Err: ctx.Err()}
}

func TestInstallTimeoutPersistsFailure(t *testing.T) {
	options := testOptions(t, contextRunner{})
	options.CommandTimeout = time.Millisecond
	result, err := Install(context.Background(), "go", options)
	if err == nil || !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("unexpected timeout result: %+v, %v", result, err)
	}
	if result.State != "failed" {
		t.Fatalf("state = %q, want failed", result.State)
	}
	recordBytes, readErr := os.ReadFile(filepath.Join(options.Paths.InstallationsDir(), "go.json"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !strings.Contains(string(recordBytes), `"state": "failed"`) {
		t.Fatalf("record did not contain failed state: %s", recordBytes)
	}
}

func TestInstallWaitsForAnotherMobdeskInstallation(t *testing.T) {
	options := testOptions(t, &fakeRunner{})
	release, err := acquireInstallLock(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	options.LockTimeout = time.Millisecond
	var updates []string
	options.Progress = func(message string) { updates = append(updates, message) }
	_, err = Install(context.Background(), "go", options)
	if err == nil || i18n.ErrorCode(err) != "install_wait" {
		t.Fatalf("unexpected lock error: %v", err)
	}
	if !slices.Equal(updates, []string{"wait for another Mobdesk installation"}) {
		t.Fatalf("progress = %v", updates)
	}
	if _, statErr := os.Stat(options.Paths.InstallationsDir()); !os.IsNotExist(statErr) {
		t.Fatalf("lock failure unexpectedly created installation state: %v", statErr)
	}
}

func TestInstallRejectsUnsupportedLanguage(t *testing.T) {
	_, err := Install(context.Background(), "rust", Options{Runner: &fakeRunner{}})
	if err == nil || i18n.ErrorCode(err) != "install_unsupported" {
		t.Fatalf("unexpected error: %v", err)
	}
}
