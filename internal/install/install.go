package install

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/ericklucioh/mobdesk/internal/paths"
)

const (
	defaultCommandTimeout = 10 * time.Minute
	defaultLockTimeout    = 5 * time.Minute
)

var catalog = []Language{
	{Name: "go", Aliases: []string{"golang"}, Package: "golang", Executable: "go", VersionArg: []string{"version"}, Kind: "language", InstallKind: "pkg"},
	{Name: "python", Aliases: []string{"python3"}, Package: "python", Executable: "python", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "pkg"},
	{Name: "node", Aliases: []string{"nodejs"}, Package: "nodejs", Executable: "node", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "pkg"},
	{Name: "c", Aliases: []string{"c-lang"}, Package: "clang", Executable: "clang", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "pkg"},
	{Name: "cpp", Aliases: []string{"c++", "cplusplus"}, Package: "clang", Executable: "clang++", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "pkg"},
	{Name: "lua", Aliases: []string{"lua5.4"}, Package: "lua", Executable: "lua", VersionArg: []string{"-v"}, Kind: "language", InstallKind: "pkg"},
	{Name: "java", Aliases: []string{"openjdk", "openjdk-21"}, Package: "openjdk-21", Executable: "java", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "pkg"},
	{Name: "kotlin", Package: "kotlin", Executable: "kotlin", VersionArg: []string{"-version"}, Kind: "language", InstallKind: "pkg", Requires: []string{"java"}},
	{Name: "gradle", Package: "gradle", Executable: "gradle", VersionArg: []string{"--version"}, Kind: "development", InstallKind: "pkg", Requires: []string{"java"}},
	{Name: "maven", Aliases: []string{"mvn"}, Package: "maven", Executable: "mvn", VersionArg: []string{"--version"}, Kind: "development", InstallKind: "pkg", Requires: []string{"java"}},
	{Name: "git", Package: "git", Executable: "git", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "pkg"},
	{Name: "gh", Aliases: []string{"github-cli"}, Package: "gh", Executable: "gh", VersionArg: []string{"--version"}, Kind: "development", InstallKind: "pkg"},
	{Name: "tmux", Package: "tmux", Executable: "tmux", VersionArg: []string{"-V"}, Kind: "terminal", InstallKind: "pkg"},
	{Name: "zellij", Package: "zellij", Executable: "zellij", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "pkg"},
	{Name: "micro", Package: "micro", Executable: "micro", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "pkg"},
	{Name: "neovim", Aliases: []string{"nvim"}, Package: "neovim", Executable: "nvim", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "pkg"},
	{Name: "lazygit", Package: "lazygit", Executable: "lazygit", VersionArg: []string{"--version"}, Kind: "development", InstallKind: "pkg"},
	{Name: "tree", Package: "tree", Executable: "tree", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "pkg"},
	{Name: "ttt", Package: "github.com/eugenioenko/ttt/cmd/ttt@v1.1.0", Executable: "ttt", VersionArg: []string{"--help"}, Kind: "development", InstallKind: "ttt", Requires: []string{"go"}},
	{Name: "htop", Package: "htop", Executable: "htop", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "pkg"},
	{Name: "ncdu", Package: "ncdu", Executable: "ncdu", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "pkg"},
	{Name: "inxi", Package: "inxi", Executable: "inxi", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "pkg"},
	{Name: "speedtest-cli", Package: "speedtest-cli", Executable: "speedtest-cli", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "pkg"},
	{Name: "posting", Package: "posting", Executable: "posting", VersionArg: []string{"--help"}, Kind: "terminal", InstallKind: "pkg"},
	{Name: "opencode-cli", Aliases: []string{"opencode"}, Package: "opencode-ai", Executable: "opencode", VersionArg: []string{"--version"}, Kind: "ai", InstallKind: "npm", Requires: []string{"node"}, UserBin: true},
	{Name: "codex-cli", Aliases: []string{"codex"}, Package: "@openai/codex", Executable: "codex", VersionArg: []string{"--version"}, Kind: "ai", InstallKind: "npm", Requires: []string{"node"}, UserBin: true},
	{Name: "claudecode-cli", Aliases: []string{"claude-code"}, Package: "@anthropic-ai/claude-code", Executable: "claude", VersionArg: []string{"--version"}, Kind: "ai", InstallKind: "npm", Requires: []string{"node"}, UserBin: true},
	{Name: "leetgo", Package: "github.com/j178/leetgo@latest", Executable: "leetgo", VersionArg: []string{"--help"}, Kind: "development", InstallKind: "go", Requires: []string{"go"}},
}

func init() {
	for index := range catalog {
		catalog[index].Targets = []Target{{Runtime: RuntimeTermux, Package: catalog[index].Package, InstallKind: catalog[index].InstallKind}}
	}
}

type Options struct {
	Paths          paths.Paths
	Runner         CommandRunner
	Now            func() time.Time
	CommandTimeout time.Duration
	LockTimeout    time.Duration
	Progress       func(string)
}

func Languages() []Language {
	result := make([]Language, len(catalog))
	copy(result, catalog)
	return result
}

func Tools() []Language { return Languages() }

func Resolve(name string) (Language, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, language := range catalog {
		if name == language.Name {
			return language, true
		}
		for _, alias := range language.Aliases {
			if name == alias {
				return language, true
			}
		}
	}
	return Language{}, false
}

func Install(ctx context.Context, name string, options Options) (Result, error) {
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.CommandTimeout <= 0 {
		options.CommandTimeout = defaultCommandTimeout
	}
	if options.LockTimeout <= 0 {
		options.LockTimeout = defaultLockTimeout
	}
	release, err := acquireInstallLock(ctx, options)
	if err != nil {
		return Result{}, err
	}
	defer release()
	return install(ctx, name, options)
}

// MigrateLegacyInstallations reinstalls records created by the Ubuntu-first
// catalog with their Termux-native targets. Failed tools remain recorded so the
// user can retry them after resolving the reported package issue.
func MigrateLegacyInstallations(ctx context.Context, options Options) ([]Result, error) {
	entries, err := os.ReadDir(options.Paths.InstallationsDir())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ler instalações legadas: %w", err)
	}
	var results []Result
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		payload, err := os.ReadFile(filepath.Join(options.Paths.InstallationsDir(), entry.Name()))
		if err != nil {
			return results, fmt.Errorf("ler instalação legada %s: %w", entry.Name(), err)
		}
		var record InstallationRecord
		if err := json.Unmarshal(payload, &record); err != nil || record.Name == "" || record.State != "installed" || record.Runtime == RuntimeTermux {
			continue
		}
		result, err := Install(ctx, record.Name, options)
		if err != nil {
			return results, fmt.Errorf("migrar %s para Termux: %w", record.Name, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func install(ctx context.Context, name string, options Options) (Result, error) {
	language, ok := Resolve(name)
	if !ok {
		return Result{}, fmt.Errorf("linguagem não suportada %q", name)
	}
	target, ok := language.Target(RuntimeTermux)
	if !ok {
		return Result{}, fmt.Errorf("%s não possui instalação nativa para Termux", language.Name)
	}
	language.Package = target.Package
	language.InstallKind = target.InstallKind
	runner := options.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	for _, prerequisite := range language.Requires {
		progress(options, fmt.Sprintf("Preparando dependência %s", prerequisite))
		if _, err := install(ctx, prerequisite, options); err != nil {
			return Result{}, fmt.Errorf("preparar dependência %s para %s: %w", prerequisite, language.Name, err)
		}
	}
	now := options.Now().UTC()
	installationsDir := options.Paths.InstallationsDir()
	logsDir := options.Paths.InstallLogsDir()
	logPath := filepath.Join(logsDir, language.Name+".log")
	result := Result{
		SchemaVersion: 1,
		Language:      language.Name,
		Package:       language.Package,
		Executable:    language.Executable,
		State:         "installing",
		LogPath:       logPath,
		Runtime:       RuntimeTermux,
	}
	record := InstallationRecord{
		Name:          language.Name,
		Kind:          language.Kind,
		Package:       language.Package,
		Executable:    language.Executable,
		State:         "installing",
		LastAttemptAt: now,
		LogPath:       logPath,
		Runtime:       RuntimeTermux,
	}
	if err := os.MkdirAll(installationsDir, 0o700); err != nil {
		return result, fmt.Errorf("criar estado da instalação: %w", err)
	}
	if err := os.MkdirAll(logsDir, 0o700); err != nil {
		return result, fmt.Errorf("criar diretório de logs da instalação: %w", err)
	}
	if err := saveRecord(installationsDir, record); err != nil {
		return result, fmt.Errorf("registrar tentativa de instalação: %w", err)
	}

	progress(options, fmt.Sprintf("Verificando %s", language.Name))
	version := runToolVersion(ctx, runner, options.CommandTimeout, logPath, language)
	if version.Err != nil {
		progress(options, "Atualizando repositórios do Termux")
		progress(options, "Aguardando gerenciador de pacotes")
		if update := runPkgLogged(ctx, runner, options.CommandTimeout, logPath, "update"); update.Err != nil {
			err := fmt.Errorf("atualizar repositórios do Termux para %s: %w", language.Name, update.Err)
			return failInstallation(installationsDir, record, result, err)
		}
		progress(options, fmt.Sprintf("Instalando %s", language.Name))
		if install := installTool(ctx, runner, options.CommandTimeout, logPath, language); install.Err != nil {
			err := fmt.Errorf("instalar %s: %w", language.Name, install.Err)
			return failInstallation(installationsDir, record, result, err)
		}
		result.Changed = true
		progress(options, fmt.Sprintf("Verificando %s após a instalação", language.Name))
		version = runToolVersion(ctx, runner, options.CommandTimeout, logPath, language)
	}
	if version.Err != nil {
		err := fmt.Errorf("verificar %s após instalação: %w", language.Name, version.Err)
		return failInstallation(installationsDir, record, result, err)
	}
	result.Version = commandOutput(version)
	result.Installed = true
	result.State = "installed"
	record.State = result.State
	record.Version = result.Version
	record.InstalledAt = options.Now().UTC()
	if err := saveRecord(installationsDir, record); err != nil {
		return result, fmt.Errorf("registrar instalação concluída: %w", err)
	}
	return result, nil
}

func progress(options Options, message string) {
	if options.Progress != nil {
		options.Progress(message)
	}
}

func acquireInstallLock(parent context.Context, options Options) (func(), error) {
	path := options.Paths.InstallLock()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("criar lock de instalação: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("abrir lock de instalação: %w", err)
	}
	ctx, cancel := context.WithTimeout(parent, options.LockTimeout)
	defer cancel()
	waited := false
	for {
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return func() {
				_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
				_ = file.Close()
			}, nil
		}
		if err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
			_ = file.Close()
			return nil, fmt.Errorf("bloquear instalação: %w", err)
		}
		if !waited {
			progress(options, "Aguardando outra instalação do Mobdesk")
			waited = true
		}
		select {
		case <-ctx.Done():
			_ = file.Close()
			return nil, fmt.Errorf("aguardar outra instalação do Mobdesk: %w", ctx.Err())
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func runToolVersion(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, tool Language) CommandResult {
	if !tool.UserBin {
		return runTermuxLogged(ctx, runner, timeout, logPath, tool.Executable, tool.VersionArg...)
	}
	args := append([]string{"-ec", `PATH="$HOME/.local/bin:$PATH"; exec "$@"`, "--", tool.Executable}, tool.VersionArg...)
	return runTermuxLogged(ctx, runner, timeout, logPath, "sh", args...)
}

func installTool(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, tool Language) CommandResult {
	switch tool.InstallKind {
	case "npm":
		return runTermuxLogged(ctx, runner, timeout, logPath, "sh", "-ec", "NPM_CONFIG_PREFIX=\"$HOME/.local\" npm install -g \"$1\"", "--", tool.Package)
	case "go":
		return runTermuxLogged(ctx, runner, timeout, logPath, "sh", "-ec", "mkdir -p \"$HOME/.local/bin\"; GOBIN=\"$HOME/.local/bin\" go install \"$1\"", "--", tool.Package)
	case "ttt":
		if result := runPkgLogged(ctx, runner, timeout, logPath, "install", "-y", "git", "ripgrep"); result.Err != nil {
			return result
		}
		return runTermuxLogged(ctx, runner, timeout, logPath, "sh", "-ec", "mkdir -p \"$HOME/.local/bin\"; GOBIN=\"$HOME/.local/bin\" go install \"$1\"", "--", tool.Package)
	case "cargo":
		if result := runPkgLogged(ctx, runner, timeout, logPath, "install", "-y", "cargo"); result.Err != nil {
			return result
		}
		return runTermuxLogged(ctx, runner, timeout, logPath, "sh", "-ec", "CARGO_INSTALL_ROOT=\"$HOME/.local\" cargo install --locked \"$1\"", "--", tool.Package)
	case "script":
		return runTermuxLogged(ctx, runner, timeout, logPath, "sh", "-ec", tool.Script)
	case "gh-extension":
		return runTermuxLogged(ctx, runner, timeout, logPath, "gh", "extension", "install", tool.Package)
	default:
		return runPkgLogged(ctx, runner, timeout, logPath, "install", "-y", tool.Package)
	}
}

func runPkgLogged(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, args ...string) CommandResult {
	return runTermuxLogged(ctx, runner, timeout, logPath, "pkg", args...)
}

func runTermuxLogged(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, name string, args ...string) CommandResult {
	commandContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	commandArgs := append([]string{name}, args...)
	result := runner.Run(commandContext, name, args...)
	_ = appendLog(logPath, commandArgs, result)
	return result
}

func appendLog(path string, args []string, result CommandResult) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if _, err := fmt.Fprintf(file, "\n$ %s\n", strings.Join(args, " ")); err != nil {
		return err
	}
	if len(result.Stdout) > 0 {
		if _, err := fmt.Fprintf(file, "[stdout]\n%s\n", result.Stdout); err != nil {
			return err
		}
	}
	if len(result.Stderr) > 0 {
		if _, err := fmt.Fprintf(file, "[stderr]\n%s\n", result.Stderr); err != nil {
			return err
		}
	}
	if result.Err != nil {
		_, err = fmt.Fprintf(file, "[error] %v\n", result.Err)
	}
	return err
}

func saveRecord(directory string, record InstallationRecord) error {
	payload, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(directory, record.Name+".json")
	temporary, err := os.CreateTemp(directory, record.Name+"-*.tmp")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(append(payload, '\n')); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryName, path)
}

func failInstallation(directory string, record InstallationRecord, result Result, installErr error) (Result, error) {
	result.State = "failed"
	record.State = result.State
	record.LastError = installErr.Error()
	if err := saveRecord(directory, record); err != nil {
		return result, fmt.Errorf("%v; registrar falha da instalação: %w", installErr, err)
	}
	return result, installErr
}

func commandOutput(result CommandResult) string {
	output := result.Stdout
	if len(output) == 0 {
		output = result.Stderr
	}
	return strings.TrimSpace(string(output))
}
