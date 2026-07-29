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
	aptLockTimeoutSeconds = 300
	ubuntuPath            = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
)

var catalog = []Language{
	{Name: "go", Aliases: []string{"golang"}, Package: "golang", Executable: "go", VersionArg: []string{"version"}, Kind: "language", InstallKind: "apt"},
	{Name: "python", Aliases: []string{"python3"}, Package: "python3", Executable: "python3", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "apt"},
	{Name: "node", Aliases: []string{"nodejs"}, Package: "nodejs", Executable: "node", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "node"},
	{Name: "c", Aliases: []string{"c-lang"}, Package: "clang", Executable: "clang", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "apt"},
	{Name: "cpp", Aliases: []string{"c++", "cplusplus"}, Package: "clang", Executable: "clang++", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "apt"},
	{Name: "lua", Aliases: []string{"lua5.4"}, Package: "lua5.4", Executable: "lua5.4", VersionArg: []string{"-v"}, Kind: "language", InstallKind: "apt"},
	{Name: "git", Package: "git", Executable: "git", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "apt"},
	{Name: "gh", Aliases: []string{"github-cli"}, Package: "gh", Executable: "gh", VersionArg: []string{"--version"}, Kind: "development", InstallKind: "apt"},
	{Name: "tmux", Package: "tmux", Executable: "tmux", VersionArg: []string{"-V"}, Kind: "terminal", InstallKind: "apt"},
	{Name: "zellij", Package: "zellij", Executable: "zellij", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "script", UserBin: true, Script: "apt-get install -y ca-certificates curl tar; mkdir -p \"$HOME/.local/bin\"; archive=$(mktemp); curl -fsSL https://github.com/zellij-org/zellij/releases/download/v0.44.3/zellij-aarch64-unknown-linux-musl.tar.gz -o \"$archive\"; printf '%s  %s\\n' '15e6534d42644d66973d136c590c49739dcfd6a1a2a0d3d917973f16c81b45fb' \"$archive\" | sha256sum -c -; tar -xzf \"$archive\" -C \"$HOME/.local/bin\" zellij; chmod 0755 \"$HOME/.local/bin/zellij\"; rm -f \"$archive\""},
	{Name: "micro", Package: "micro", Executable: "micro", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "apt"},
	{Name: "lazygit", Package: "github.com/jesseduffield/lazygit@v0.63.1", Executable: "lazygit", VersionArg: []string{"--version"}, Kind: "development", InstallKind: "script", Script: githubReleaseScript("jesseduffield/lazygit", "v0.63.1", "lazygit_0.63.1_linux_arm64.tar.gz", "555dbc9a8efcf2e33bc24e7fbd9463e9fa375e3c5e23cc270763733c38eeae36", "lazygit_0.63.1_linux_x86_64.tar.gz", "8e033bc78c8e192dee9510e951f6c9e154289b7198d22c924ed1d0a951b0dac1", "lazygit")},
	{Name: "tree", Package: "tree", Executable: "tree", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "apt"},
	{Name: "ttt", Package: "github.com/eugenioenko/ttt/cmd/ttt@v1.1.0", Executable: "ttt", VersionArg: []string{"--help"}, Kind: "development", InstallKind: "ttt", Requires: []string{"go"}},
	{Name: "htop", Package: "htop", Executable: "htop", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "apt"},
	{Name: "ncdu", Package: "ncdu", Executable: "ncdu", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "apt"},
	{Name: "inxi", Package: "inxi", Executable: "inxi", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "apt"},
	{Name: "speedtest-cli", Package: "speedtest-cli", Executable: "speedtest-cli", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "apt"},
	{Name: "posting", Package: "posting", Executable: "posting", VersionArg: []string{"--help"}, Kind: "terminal", InstallKind: "pipx", Requires: []string{"python"}},
	{Name: "yazi", Aliases: []string{"yazi-fm"}, Package: "yazi@v26.5.6", Executable: "yazi", VersionArg: []string{"--version"}, Kind: "file", InstallKind: "script", UserBin: true, Script: yaziReleaseScript()},
	{Name: "tuifi", Aliases: []string{"tuifimanager"}, Package: "TUIFIManager==5.2.6", Executable: "tuifi", VersionArg: []string{"--version"}, Kind: "file", InstallKind: "script", Requires: []string{"python"}, Script: tuifiInstallScript()},
	{Name: "opencode-cli", Aliases: []string{"opencode"}, Package: "opencode-ai", Executable: "opencode", VersionArg: []string{"--version"}, Kind: "ai", InstallKind: "npm", Requires: []string{"node"}, UserBin: true},
	{Name: "codex-cli", Aliases: []string{"codex"}, Package: "@openai/codex", Executable: "codex", VersionArg: []string{"--version"}, Kind: "ai", InstallKind: "npm", Requires: []string{"node"}, UserBin: true},
	{Name: "claudecode-cli", Aliases: []string{"claude-code"}, Package: "@anthropic-ai/claude-code", Executable: "claude", VersionArg: []string{"--version"}, Kind: "ai", InstallKind: "npm", Requires: []string{"node"}, UserBin: true},
	{Name: "leetgo", Package: "github.com/j178/leetgo@v1.4.17", Executable: "leetgo", VersionArg: []string{"--help"}, Kind: "development", InstallKind: "script", Script: githubReleaseScript("j178/leetgo", "v1.4.17", "leetgo_linux_arm64.tar.gz", "de77054553b61c1733f9b034e4a976630a3da585e414f93f0ce13ada5dd80ca4", "leetgo_linux_x86_64.tar.gz", "fe18dc54f2784aded76ef1e04e6917d6d9d8731520bbe232328ba942b5b3c47b", "leetgo")},
}

func githubReleaseScript(repository, version, arm64Archive, arm64Checksum, amd64Archive, amd64Checksum, executable string) string {
	return fmt.Sprintf(`set -eu
apt-get -o DPkg::Lock::Timeout=300 install -y ca-certificates curl tar
case "$(uname -m)" in
aarch64|arm64)
    archive_name=%q
    checksum=%q
    ;;
x86_64|amd64)
    archive_name=%q
    checksum=%q
    ;;
*)
    printf 'arquitetura nao suportada: %%s\n' "$(uname -m)" >&2
    exit 1
    ;;
esac
archive=$(mktemp)
trap 'rm -f "$archive"' EXIT
curl -fsSL "https://github.com/%s/releases/download/%s/$archive_name" -o "$archive"
printf '%%s  %%s\n' "$checksum" "$archive" | sha256sum -c -
tar -xzf "$archive" -C /usr/local/bin %s
chmod 0755 "/usr/local/bin/%s"`, arm64Archive, arm64Checksum, amd64Archive, amd64Checksum, repository, version, executable, executable)
}

func yaziReleaseScript() string {
	return `set -eu
apt-get -o DPkg::Lock::Timeout=300 install -y ca-certificates curl unzip file ffmpeg 7zip jq poppler-utils fd-find ripgrep fzf zoxide chafa imagemagick
case "$(uname -m)" in
aarch64|arm64)
    archive_name=yazi-aarch64-unknown-linux-gnu.zip
    checksum=c38b07961e7fc4c76503fd0f4a1b4bd0b379a99835b818cd899b0315c728e1e1
    ;;
x86_64|amd64)
    archive_name=yazi-x86_64-unknown-linux-gnu.zip
    checksum=1c9096f0a83b8102c194385f644cdeff93cc8269426163c9d033041ebd537bd2
    ;;
*)
    printf 'arquitetura nao suportada: %s\n' "$(uname -m)" >&2
    exit 1
    ;;
esac
archive=$(mktemp)
temporary=$(mktemp -d)
trap 'rm -f "$archive"; rm -rf "$temporary"' EXIT
curl -fsSL "https://github.com/sxyazi/yazi/releases/download/v26.5.6/$archive_name" -o "$archive"
printf '%s  %s\n' "$checksum" "$archive" | sha256sum -c -
unzip -q "$archive" -d "$temporary"
yazi_binary=$(find "$temporary" -type f -name yazi -print -quit)
ya_binary=$(find "$temporary" -type f -name ya -print -quit)
test -n "$yazi_binary" -a -n "$ya_binary"
mkdir -p "$HOME/.local/bin"
install -m 0755 "$yazi_binary" "$HOME/.local/bin/yazi"
install -m 0755 "$ya_binary" "$HOME/.local/bin/ya"`
}

func tuifiInstallScript() string {
	return `set -eu
apt-get -o DPkg::Lock::Timeout=300 install -y build-essential python3-dev libncurses-dev pipx
PIPX_BIN_DIR=/usr/local/bin pipx install --force TUIFIManager==5.2.6`
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

func install(ctx context.Context, name string, options Options) (Result, error) {
	language, ok := Resolve(name)
	if !ok {
		return Result{}, fmt.Errorf("linguagem não suportada %q", name)
	}
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
	}
	record := InstallationRecord{
		Name:          language.Name,
		Kind:          language.Kind,
		Package:       language.Package,
		Executable:    language.Executable,
		State:         "installing",
		LastAttemptAt: now,
		LogPath:       logPath,
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
		progress(options, "Atualizando índices do Ubuntu")
		progress(options, "Aguardando gerenciador de pacotes")
		if update := runAptLogged(ctx, runner, options.CommandTimeout, logPath, "update"); update.Err != nil {
			err := fmt.Errorf("atualizar índices do Ubuntu para %s: %w", language.Name, update.Err)
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
		return runUbuntuLogged(ctx, runner, timeout, logPath, tool.Executable, tool.VersionArg...)
	}
	args := append([]string{"-ec", `PATH="$HOME/.local/bin:$PATH"; exec "$@"`, "--", tool.Executable}, tool.VersionArg...)
	return runUbuntuLogged(ctx, runner, timeout, logPath, "sh", args...)
}

func installTool(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, tool Language) CommandResult {
	switch tool.InstallKind {
	case "node":
		return runAptLogged(ctx, runner, timeout, logPath, "install", "-y", "nodejs", "npm")
	case "npm":
		if result := runAptLogged(ctx, runner, timeout, logPath, "install", "-y", "npm"); result.Err != nil {
			return result
		}
		return runUbuntuLogged(ctx, runner, timeout, logPath, "env", "NPM_CONFIG_PREFIX=/root/.local", "npm", "install", "-g", tool.Package)
	case "pipx":
		if result := runAptLogged(ctx, runner, timeout, logPath, "install", "-y", "pipx"); result.Err != nil {
			return result
		}
		return runUbuntuLogged(ctx, runner, timeout, logPath, "env", "PIPX_BIN_DIR=/usr/local/bin", "pipx", "install", tool.Package)
	case "go":
		return runUbuntuLogged(ctx, runner, timeout, logPath, "env", "GOBIN=/usr/local/bin", "go", "install", tool.Package)
	case "ttt":
		if result := runAptLogged(ctx, runner, timeout, logPath, "install", "-y", "git", "ripgrep"); result.Err != nil {
			return result
		}
		return runUbuntuLogged(ctx, runner, timeout, logPath, "env", "GOBIN=/usr/local/bin", "go", "install", tool.Package)
	case "cargo":
		if result := runAptLogged(ctx, runner, timeout, logPath, "install", "-y", "cargo"); result.Err != nil {
			return result
		}
		return runUbuntuLogged(ctx, runner, timeout, logPath, "env", "CARGO_INSTALL_ROOT=/usr/local", "cargo", "install", "--locked", tool.Package)
	case "script":
		return runUbuntuLogged(ctx, runner, timeout, logPath, "sh", "-ec", tool.Script)
	case "gh-extension":
		return runUbuntuLogged(ctx, runner, timeout, logPath, "gh", "extension", "install", tool.Package)
	default:
		return runAptLogged(ctx, runner, timeout, logPath, "install", "-y", tool.Package)
	}
}

func runAptLogged(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, args ...string) CommandResult {
	aptArgs := append([]string{"-o", fmt.Sprintf("DPkg::Lock::Timeout=%d", aptLockTimeoutSeconds)}, args...)
	return runUbuntuLogged(ctx, runner, timeout, logPath, "apt-get", aptArgs...)
}

func runUbuntuLogged(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, name string, args ...string) CommandResult {
	commandContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	loginArgs := append([]string{"login", "ubuntu", "--", "env", "PATH=" + ubuntuPath, name}, args...)
	result := runner.Run(commandContext, "proot-distro", loginArgs...)
	_ = appendLog(logPath, loginArgs, result)
	return result
}

func appendLog(path string, args []string, result CommandResult) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if _, err := fmt.Fprintf(file, "\n$ proot-distro %s\n", strings.Join(args, " ")); err != nil {
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
