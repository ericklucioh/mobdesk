package install

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

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
	for _, name := range []string{"git", "gh", "tmux", "zellij", "micro", "lazygit", "tree", "ttt", "posting", "btop", "ncdu", "inxi", "speedtest-cli", "opencode-cli", "codex-cli", "claudecode-cli", "leetgo"} {
		tool, ok := Resolve(name)
		if !ok || tool.Kind == "" || tool.InstallKind == "" {
			t.Fatalf("Resolve(%q) = %+v, %t", name, tool, ok)
		}
	}
	for _, unsupported := range []string{"tuifi", "clin", "glint"} {
		if _, ok := Resolve(unsupported); ok {
			t.Fatalf("Resolve(%q) unexpectedly succeeded", unsupported)
		}
	}
}

func TestInstallToolUsesPrivateNPMPrefix(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- apt-get install -y npm":                                        {{}},
		"proot-distro login ubuntu -- env NPM_CONFIG_PREFIX=/root/.local npm install -g opencode-ai": {{}},
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

func TestInstallNodeInstallsNPM(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- apt-get install -y nodejs npm": {{}},
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
		"proot-distro login ubuntu -- apt-get install -y git ripgrep":                                                {{}},
		"proot-distro login ubuntu -- env GOBIN=/usr/local/bin go install github.com/eugenioenko/ttt/cmd/ttt@v1.1.0": {{}},
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
		"proot-distro login ubuntu -- sh -ec PATH=\"$HOME/.local/bin:$PATH\"; exec \"$@\" -- zellij --version": {{Stdout: []byte("zellij 0.44.3")}},
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
		"proot-distro login ubuntu -- python3 --version": {{Stdout: []byte("Python 3.12.1\n")}},
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
		"proot-distro login ubuntu -- go version":                {{Err: errors.New("not found")}, {Stdout: []byte("go version go1.26.5 linux/arm64\n")}},
		"proot-distro login ubuntu -- apt-get update":            {{}},
		"proot-distro login ubuntu -- apt-get install -y golang": {{}},
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
	wantUpdates := []string{"Verificando go", "Atualizando índices do Ubuntu", "Instalando go", "Verificando go após a instalação"}
	if !slices.Equal(updates, wantUpdates) {
		t.Fatalf("progress = %v, want %v", updates, wantUpdates)
	}
}

func TestInstallPersistsRecordAndCommandLog(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- node --version": {{Stdout: []byte("v22.1.0\n")}},
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

func TestInstallRejectsUnsupportedLanguage(t *testing.T) {
	_, err := Install(context.Background(), "rust", Options{Runner: &fakeRunner{}})
	if err == nil || !strings.Contains(err.Error(), "não suportada") {
		t.Fatalf("unexpected error: %v", err)
	}
}
