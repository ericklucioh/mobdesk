package install

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/paths"
)

const (
	defaultCommandTimeout = 10 * time.Minute
	defaultLockTimeout    = 5 * time.Minute
	aptLockTimeoutSeconds = 300
	ubuntuPath            = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
	storageWarningBytes   = 20 * 1024 * 1024 * 1024
	storageBlockBytes     = 10 * 1024 * 1024 * 1024
)

var catalog = []AppProfile{
	{Name: "go", Aliases: []string{"golang"}, DescriptionID: i18n.AppGoDescription, Usage: "go [command]", Package: "golang", Executable: "go", VersionArg: []string{"version"}, Kind: "language", InstallKind: "apt", StorageEstimate: plannedStorage(180, 300, 0, 50, 0, 5)},
	{Name: "java", Aliases: []string{"openjdk"}, DescriptionID: i18n.AppJavaDescription, Usage: "java [options] <class>", Package: "openjdk-21-jdk", Executable: "java", RequiredExecutables: []ExecutableSpec{{Name: "java", VersionArg: []string{"--version"}}, {Name: "javac", VersionArg: []string{"--version"}}, {Name: "jar", VersionArg: []string{"--version"}}}, VersionArg: []string{"--version"}, Kind: "language", InstallKind: "apt", StorageEstimate: plannedStorage(160, 340, 0, 0, 0, 5)},
	{Name: "kotlin", Aliases: []string{"kotlin-jvm", "kotlinc"}, DescriptionID: i18n.AppKotlinDescription, Usage: "kotlinc [options] <source files>", Package: "kotlin-compiler-2.2.20", Executable: "kotlinc", RequiredExecutables: []ExecutableSpec{{Name: "kotlinc", VersionArg: []string{"-version"}}, {Name: "kotlin", VersionArg: []string{"-version"}}}, VersionArg: []string{"-version"}, Kind: "language", InstallKind: "script", Requires: []string{"java"}, UserBin: true, Script: kotlinCompilerInstallScript(), StorageEstimate: plannedStorage(85, 110, 0, 0, 1, 5)},
	{Name: "gradle", Aliases: []string{"gradle-build"}, DescriptionID: i18n.AppGradleDescription, Usage: "gradle [options] [tasks...]", Package: "gradle", Executable: "gradle", VersionArg: []string{"--version"}, Kind: "build", InstallKind: "apt", Requires: []string{"java"}, StorageEstimate: plannedStorage(45, 90, 0, 20, 1, 5)},
	{Name: "maven", Aliases: []string{"mvn"}, DescriptionID: i18n.AppMavenDescription, Usage: "mvn [options] [goals...]", Package: "maven", Executable: "mvn", VersionArg: []string{"--version"}, Kind: "build", InstallKind: "apt", Requires: []string{"java"}, StorageEstimate: plannedStorage(15, 35, 0, 10, 1, 5)},
	{Name: "python", Aliases: []string{"python3"}, DescriptionID: i18n.AppPythonDescription, Usage: "python3 [script]", Package: "python3", Executable: "python3", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "apt", StorageEstimate: plannedStorage(35, 60, 0, 20, 0, 5)},
	{Name: "node", Aliases: []string{"nodejs"}, DescriptionID: i18n.AppNodeDescription, Usage: "node [script]", Package: "nodejs", Executable: "node", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "node", StorageEstimate: plannedStorage(70, 130, 20, 60, 0, 10)},
	{Name: "c", Aliases: []string{"c-lang"}, DescriptionID: i18n.AppCDescription, Usage: "clang [options] <files...>", Package: "clang", Executable: "clang", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "apt", StorageEstimate: plannedStorage(250, 450, 20, 80, 0, 10)},
	{Name: "cpp", Aliases: []string{"c++", "cplusplus"}, DescriptionID: i18n.AppCPPDescription, Usage: "clang++ [options] <files...>", Package: "clang", Executable: "clang++", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "apt", StorageEstimate: plannedStorage(250, 450, 20, 80, 0, 10)},
	{Name: "lua", Aliases: []string{"lua5.4"}, DescriptionID: i18n.AppLuaDescription, Usage: "lua5.4 [script]", Package: "lua5.4", Executable: "lua5.4", VersionArg: []string{"-v"}, Kind: "language", InstallKind: "apt", StorageEstimate: plannedStorage(2, 6, 0, 5, 0, 2)},
	{Name: "git", DescriptionID: i18n.AppGitDescription, Usage: "git <command> [options]", Package: "git", Executable: "git", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "apt", StorageEstimate: plannedStorage(35, 60, 0, 10, 0, 5)},
	{Name: "gh", Aliases: []string{"github-cli"}, DescriptionID: i18n.AppGHDescription, Usage: "gh <command> [options]", Package: "gh", Executable: "gh", VersionArg: []string{"--version"}, Kind: "development", InstallKind: "apt", StorageEstimate: plannedStorage(30, 50, 0, 10, 0, 5)},
	{Name: "tmux", DescriptionID: i18n.AppTmuxDescription, Usage: "tmux [command]", Package: "tmux", Executable: "tmux", VersionArg: []string{"-V"}, Kind: "terminal", InstallKind: "apt", StorageEstimate: plannedStorage(2, 5, 0, 2, 0, 2)},
	{Name: "zellij", DescriptionID: i18n.AppZellijDescription, Usage: "zellij [options]", Package: "zellij", Executable: "zellij", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "script", UserBin: true, Script: "apt-get install -y ca-certificates curl tar; mkdir -p \"$HOME/.local/bin\"; archive=$(mktemp); curl -fsSL https://github.com/zellij-org/zellij/releases/download/v0.44.3/zellij-aarch64-unknown-linux-musl.tar.gz -o \"$archive\"; printf '%s  %s\\n' '15e6534d42644d66973d136c590c49739dcfd6a1a2a0d3d917973f16c81b45fb' \"$archive\" | sha256sum -c -; tar -xzf \"$archive\" -C \"$HOME/.local/bin\" zellij; chmod 0755 \"$HOME/.local/bin/zellij\"; rm -f \"$archive\"", StorageEstimate: plannedStorage(20, 30, 0, 5, 0, 5)},
	{Name: "micro", DescriptionID: i18n.AppMicroDescription, Usage: "micro [files...]", Package: "micro", Executable: "micro", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "apt", StorageEstimate: plannedStorage(15, 25, 0, 5, 0, 5)},
	{Name: "lazygit", DescriptionID: i18n.AppLazygitDescription, Usage: "lazygit [directory]", Package: "github.com/jesseduffield/lazygit@v0.63.1", Executable: "lazygit", VersionArg: []string{"--version"}, Kind: "development", InstallKind: "script", Script: githubReleaseScript("jesseduffield/lazygit", "v0.63.1", "lazygit_0.63.1_linux_arm64.tar.gz", "555dbc9a8efcf2e33bc24e7fbd9463e9fa375e3c5e23cc270763733c38eeae36", "lazygit_0.63.1_linux_x86_64.tar.gz", "8e033bc78c8e192dee9510e951f6c9e154289b7198d22c924ed1d0a951b0dac1", "lazygit"), StorageEstimate: plannedStorage(15, 25, 0, 5, 0, 5)},
	{Name: "tree", DescriptionID: i18n.AppTreeDescription, Usage: "tree [directory]", Package: "tree", Executable: "tree", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "apt", StorageEstimate: plannedStorage(1, 1, 0, 1, 0, 1)},
	{Name: "ttt", DescriptionID: i18n.AppTTTDescription, Usage: "ttt [arquivos/pastas/URLs...]", Package: "github.com/eugenioenko/ttt/cmd/ttt@v1.1.0", Executable: "ttt", VersionArg: []string{"--help"}, CatalogVersion: "v1.1.0", Kind: "development", InstallKind: "ttt", Requires: []string{"go"}, StorageEstimate: plannedStorage(10, 20, 0, 10, 0, 2)},
	{Name: "htop", DescriptionID: i18n.AppHtopDescription, Usage: "htop", Package: "htop", Executable: "htop", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "apt", StorageEstimate: plannedStorage(1, 3, 0, 1, 0, 1)},
	{Name: "ncdu", DescriptionID: i18n.AppNcduDescription, Usage: "ncdu [directory]", Package: "ncdu", Executable: "ncdu", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "apt", StorageEstimate: plannedStorage(1, 3, 0, 1, 0, 1)},
	{Name: "inxi", DescriptionID: i18n.AppInxiDescription, Usage: "inxi [options]", Package: "inxi", Executable: "inxi", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "apt", StorageEstimate: plannedStorage(5, 15, 0, 5, 0, 2)},
	{Name: "speedtest-cli", DescriptionID: i18n.AppSpeedtestDescription, Usage: "speedtest-cli", Package: "speedtest-cli", Executable: "speedtest-cli", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "apt", StorageEstimate: plannedStorage(5, 15, 0, 10, 0, 2)},
	{Name: "posting", DescriptionID: i18n.AppPostingDescription, Usage: "posting", Package: "posting", Executable: "posting", VersionArg: []string{"--help"}, Kind: "terminal", InstallKind: "pipx", Requires: []string{"python"}, StorageEstimate: plannedStorage(20, 60, 10, 40, 0, 5)},
	{Name: "yazi", Aliases: []string{"yazi-fm"}, DescriptionID: i18n.AppYaziDescription, Usage: "yazi [diret\u00f3rio]", Package: "yazi@v26.5.6", Executable: "yazi", VersionArg: []string{"--version"}, CatalogVersion: "v26.5.6", Kind: "file", InstallKind: "script", UserBin: true, Script: yaziReleaseScript(), StorageEstimate: plannedStorage(25, 40, 300, 550, 1, 20)},
	{Name: "tuifi", Aliases: []string{"tuifimanager"}, DescriptionID: i18n.AppTuifiDescription, Usage: "tuifi [diret\u00f3rio]", Package: "TUIFIManager==5.2.6", Executable: "tuifi", VersionArg: []string{"--version"}, Kind: "file", InstallKind: "script", Requires: []string{"python"}, Script: tuifiInstallScript(), StorageEstimate: plannedStorage(20, 40, 90, 180, 1, 5)},
	{Name: "neovim", Aliases: []string{"nvim"}, DescriptionID: i18n.AppNeovimDescription, Usage: "nvim [arquivo ou diret\u00f3rio]", Package: "neovim", Executable: "nvim", VersionArg: []string{"--version"}, Kind: "editor", InstallKind: "apt", ConfigProfile: "lazyvim", ConfigTarget: "/root/.config/nvim", MinimumVersion: "0.9.0", ProfileVersion: "1", StorageEstimate: plannedStorage(15, 30, 0, 20, 0, 2)},
	{Name: "opencode-cli", Aliases: []string{"opencode"}, DescriptionID: i18n.AppOpencodeDescription, Usage: "opencode [options]", Package: "opencode-ai", Executable: "opencode", VersionArg: []string{"--version"}, Kind: "ai", InstallKind: "npm", Requires: []string{"node"}, UserBin: true, StorageEstimate: plannedStorage(60, 150, 0, 100, 5, 30)},
	{Name: "codex-cli", Aliases: []string{"codex"}, DescriptionID: i18n.AppCodexDescription, Usage: "codex [options]", Package: "@openai/codex", Executable: "codex", VersionArg: []string{"--version"}, Kind: "ai", InstallKind: "npm", Requires: []string{"node"}, UserBin: true, StorageEstimate: plannedStorage(60, 150, 0, 100, 5, 30)},
	{Name: "claudecode-cli", Aliases: []string{"claude-code"}, DescriptionID: i18n.AppClaudeDescription, Usage: "claude [options]", Package: "@anthropic-ai/claude-code", Executable: "claude", VersionArg: []string{"--version"}, Kind: "ai", InstallKind: "npm", Requires: []string{"node"}, UserBin: true, StorageEstimate: plannedStorage(80, 200, 0, 100, 5, 30)},
	{Name: "leetgo", DescriptionID: i18n.AppLeetgoDescription, Usage: "leetgo [command]", Package: "github.com/j178/leetgo@v1.4.17", Executable: "leetgo", VersionArg: []string{"--help"}, CatalogVersion: "v1.4.17", Kind: "development", InstallKind: "script", Script: githubReleaseScript("j178/leetgo", "v1.4.17", "leetgo_linux_arm64.tar.gz", "de77054553b61c1733f9b034e4a976630a3da585e414f93f0ce13ada5dd80ca4", "leetgo_linux_x86_64.tar.gz", "fe18dc54f2784aded76ef1e04e6917d6d9d8731520bbe232328ba942b5b3c47b", "leetgo"), StorageEstimate: plannedStorage(10, 20, 0, 20, 0, 5)},
}

var catalogEstimateMeasuredAt = time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC)

func plannedStorage(appMin, appMax, dependenciesMin, dependenciesMax, configMin, configMax int64) *StorageEstimate {
	return &StorageEstimate{
		AppMinMB:          appMin,
		AppMaxMB:          appMax,
		DependenciesMinMB: dependenciesMin,
		DependenciesMaxMB: dependenciesMax,
		ConfigMinMB:       configMin,
		ConfigMaxMB:       configMax,
		Source:            "planning",
		Version:           "mvp-1",
		Architecture:      "arm64",
		MeasuredAt:        catalogEstimateMeasuredAt,
	}
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
    printf 'unsupported architecture: %%s\n' "$(uname -m)" >&2
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
    printf 'unsupported architecture: %s\n' "$(uname -m)" >&2
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

func kotlinCompilerInstallScript() string {
	return `set -eu
case "$(uname -m)" in
    aarch64|arm64|x86_64|amd64) ;;
    *) printf 'unsupported architecture: %s\n' "$(uname -m)" >&2; exit 1 ;;
esac
apt-get -o DPkg::Lock::Timeout=300 install -y ca-certificates curl unzip
version=2.2.20
archive=$(mktemp)
temporary=$(mktemp -d)
target="$HOME/.local/share/mobdesk/kotlin/$version"
trap 'rm -f "$archive"; rm -rf "$temporary"' EXIT
curl --proto '=https' --tlsv1.2 -fsSL "https://github.com/JetBrains/kotlin/releases/download/v$version/kotlin-compiler-$version.zip" -o "$archive"
printf '%s  %s\n' '81f0264c9073b5cbbdb3ff8418cf2c5dac076879fc156fa1a6462f5a5acc4420' "$archive" | sha256sum -c -
unzip -q "$archive" -d "$temporary"
test -x "$temporary/kotlinc/bin/kotlinc"
test -x "$temporary/kotlinc/bin/kotlin"
mkdir -p "$(dirname "$target")" "$HOME/.local/bin"
rm -rf "$target"
mv "$temporary/kotlinc" "$target"
ln -sfn "$target/bin/kotlinc" "$HOME/.local/bin/kotlinc"
ln -sfn "$target/bin/kotlin" "$HOME/.local/bin/kotlin"`
}

type Options struct {
	Paths          paths.Paths
	Runner         CommandRunner
	Interactive    bool
	Now            func() time.Time
	CommandTimeout time.Duration
	LockTimeout    time.Duration
	Progress       func(string)
	ConfigProfiles map[string]ConfigProfile
	Localizer      i18n.Localizer
	StorageFree    func(string) (int64, error)
}

func Languages(localizers ...i18n.Localizer) []AppProfile {
	result := make([]AppProfile, len(catalog))
	copy(result, catalog)
	// Keep the existing TUI presentation until its dedicated localization phase.
	localizer := i18n.New(i18n.LocalePTBR)
	if len(localizers) > 0 {
		localizer = localizers[0]
	}
	for index := range result {
		result[index].Description = localizer.Text(result[index].DescriptionID, nil)
	}
	return result
}

func Tools(localizers ...i18n.Localizer) []AppProfile { return Languages(localizers...) }

func Resolve(name string) (AppProfile, bool) {
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
	return AppProfile{}, false
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
		return Result{}, i18n.NewError(i18n.ServiceInstallUnsupported, "install_unsupported", map[string]any{"Name": name}, nil)
	}
	free, storageErr := availableStorage(options)
	if storageErr != nil {
		return Result{SchemaVersion: 1, Language: language.Name, State: "failed", StorageEstimate: language.StorageEstimate}, i18n.NewError(i18n.ServiceInstallStorage, "install_storage_check", map[string]any{"Detail": storageErr.Error()}, storageErr)
	}
	storageWarning := free < storageWarningBytes
	if free < storageBlockBytes {
		return Result{
			SchemaVersion:    1,
			Language:         language.Name,
			Package:          language.Package,
			Packages:         profilePackages(language),
			Executable:       language.Executable,
			Executables:      profileExecutables(language),
			State:            "blocked",
			Source:           "mobdesk",
			StorageEstimate:  language.StorageEstimate,
			StorageFreeBytes: free,
			StorageBlocked:   true,
		}, i18n.NewError(i18n.ServiceInstallStorage, "install_storage_blocked", map[string]any{"Name": language.Name, "Free": free}, nil)
	}
	runner := runnerFor(options)
	for _, prerequisite := range language.Requires {
		progress(options, i18n.ServiceInstallDependency, map[string]any{"Dependency": prerequisite, "Name": language.Name})
		if _, err := install(ctx, prerequisite, options); err != nil {
			return Result{}, i18n.NewError(i18n.ServiceInstallDependency, "install_dependency", map[string]any{"Dependency": prerequisite, "Name": language.Name}, err)
		}
	}
	now := options.Now().UTC()
	installationsDir := options.Paths.InstallationsDir()
	logsDir := options.Paths.InstallLogsDir()
	logPath := filepath.Join(logsDir, language.Name+".log")
	result := Result{
		SchemaVersion:    1,
		Language:         language.Name,
		Package:          language.Package,
		Packages:         profilePackages(language),
		Executable:       language.Executable,
		Executables:      profileExecutables(language),
		State:            "installing",
		LogPath:          logPath,
		Source:           "mobdesk",
		StorageEstimate:  language.StorageEstimate,
		StorageFreeBytes: free,
		StorageWarning:   storageWarning,
	}
	record := InstallationRecord{
		Name:                language.Name,
		Kind:                language.Kind,
		Package:             language.Package,
		Packages:            profilePackages(language),
		Executable:          language.Executable,
		RequiredExecutables: profileExecutables(language),
		Strategy:            language.InstallKind,
		Dependencies:        append([]string(nil), language.Requires...),
		InstalledPackages:   declaredInstalledPackages(language),
		InstalledFiles:      declaredInstalledFiles(language),
		State:               "installing",
		Source:              "mobdesk",
		LastAttemptAt:       now,
		LogPath:             logPath,
	}
	if err := os.MkdirAll(installationsDir, 0o700); err != nil {
		return result, i18n.NewError(i18n.ServiceInstallState, "install_state", nil, err)
	}
	if err := os.MkdirAll(logsDir, 0o700); err != nil {
		return result, i18n.NewError(i18n.ServiceInstallLogs, "install_logs", nil, err)
	}
	if err := saveRecord(installationsDir, record); err != nil {
		return result, i18n.NewError(i18n.ServiceInstallRecord, "install_record", nil, err)
	}

	progress(options, i18n.ServiceInstallVerify, map[string]any{"Name": language.Name})
	versions := runToolVersions(ctx, runner, options.CommandTimeout, logPath, language)
	if versionErr := firstCommandError(versions); versionErr != nil {
		progress(options, i18n.ServiceInstallRepair, map[string]any{"Name": language.Name})
		if repair := repairDpkg(ctx, runner, options.CommandTimeout, logPath); repair.Err != nil {
			err := i18n.NewError(i18n.ServiceInstallRepair, "install_repair", map[string]any{"Name": language.Name}, repair.Err)
			return failInstallation(installationsDir, record, result, err)
		}
		progress(options, i18n.ServiceInstallUpdate, map[string]any{"Name": language.Name})
		progress(options, i18n.ServiceInstallLock, nil)
		if update := runAptLogged(ctx, runner, options.CommandTimeout, logPath, "update"); update.Err != nil {
			err := i18n.NewError(i18n.ServiceInstallUpdate, "install_update", map[string]any{"Name": language.Name}, update.Err)
			return failInstallation(installationsDir, record, result, err)
		}
		progress(options, i18n.ServiceInstallTool, map[string]any{"Name": language.Name})
		if install := installTool(ctx, runner, options.CommandTimeout, logPath, language); install.Err != nil {
			err := i18n.NewError(i18n.ServiceInstallTool, "install_tool", map[string]any{"Name": language.Name}, install.Err)
			return failInstallation(installationsDir, record, result, err)
		}
		result.Changed = true
		progress(options, i18n.ServiceInstallVerify, map[string]any{"Name": language.Name})
		versions = runToolVersions(ctx, runner, options.CommandTimeout, logPath, language)
	}
	if versionErr := firstCommandError(versions); versionErr != nil {
		err := i18n.NewError(i18n.ServiceInstallVerify, "install_verify", map[string]any{"Name": language.Name}, versionErr)
		return failInstallation(installationsDir, record, result, err)
	}
	result.Version = commandVersions(language, versions)
	result.Installed = true
	result.State = "installed"
	record.State = result.State
	record.Version = result.Version
	record.InstalledAt = options.Now().UTC()
	if len(record.InstalledFiles) > 0 {
		if hashes, hashErr := hashInstalledFiles(ctx, runner, options.CommandTimeout, logPath, record.InstalledFiles); hashErr == nil {
			record.InstalledFileHashes = hashes
		}
	}
	if err := saveRecord(installationsDir, record); err != nil {
		return result, i18n.NewError(i18n.ServiceInstallRecord, "install_record", nil, err)
	}
	return result, nil
}

func availableStorage(options Options) (int64, error) {
	if options.StorageFree != nil {
		return options.StorageFree(options.Paths.Home)
	}
	if os.Getenv("MOBDESK_TEST_MODE") == "1" {
		if value, err := strconv.ParseInt(os.Getenv("MOBDESK_TEST_STORAGE_FREE_BYTES"), 10, 64); err == nil && value >= 0 {
			return value, nil
		}
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(options.Paths.Home, &stat); err != nil {
		return 0, err
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}

func progress(options Options, id i18n.MessageID, data map[string]any) {
	if options.Progress != nil {
		if data == nil {
			data = map[string]any{}
		}
		if _, ok := data["Detail"]; !ok {
			data["Detail"] = ""
		}
		localizer := options.Localizer
		if localizer.Locale == "" {
			localizer = i18n.New(i18n.LocaleENUS)
		}
		options.Progress(localizer.Text(id, data))
	}
}

func declaredInstalledPackages(profile AppProfile) []string {
	return profilePackages(profile)
}

func declaredInstalledFiles(profile AppProfile) []string {
	executables := profileExecutables(profile)
	if len(executables) == 0 {
		return nil
	}
	binDir := "/usr/local/bin"
	if profile.UserBin {
		binDir = "/root/.local/bin"
	}
	files := make([]string, 0, len(executables))
	for _, executable := range executables {
		files = append(files, filepath.Join(binDir, executable.Name))
	}
	if profile.Name == "yazi" && profile.UserBin && !containsExecutable(executables, "ya") {
		files = append(files, filepath.Join(binDir, "ya"))
	}
	if profile.InstallKind == "apt" {
		return nil
	}
	return files
}

func hashInstalledFiles(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, files []string) (map[string]string, error) {
	args := []string{"--"}
	args = append(args, files...)
	result := runUbuntuLogged(ctx, runner, timeout, logPath, "sha256sum", args...)
	if result.Err != nil {
		return nil, result.Err
	}
	hashes := make(map[string]string, len(files))
	for _, line := range strings.Split(string(result.Stdout), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		hashes[fields[1]] = fields[0]
	}
	if len(hashes) != len(files) {
		return nil, i18n.NewError(i18n.ServiceInstallHash, "install_hash", nil, nil)
	}
	return hashes, nil
}

func acquireInstallLock(parent context.Context, options Options) (func(), error) {
	path := options.Paths.InstallLock()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, i18n.NewError(i18n.ServiceInstallLock, "install_lock", nil, err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, i18n.NewError(i18n.ServiceInstallLock, "install_lock", nil, err)
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
			return nil, i18n.NewError(i18n.ServiceInstallLock, "install_lock", nil, err)
		}
		if !waited {
			progress(options, i18n.ServiceInstallWait, nil)
			waited = true
		}
		select {
		case <-ctx.Done():
			_ = file.Close()
			return nil, i18n.NewError(i18n.ServiceInstallWait, "install_wait", nil, ctx.Err())
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func runToolVersion(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, tool AppProfile) CommandResult {
	executables := profileExecutables(tool)
	if len(executables) == 0 {
		return CommandResult{Err: fmt.Errorf("profile %s has no required executable", tool.Name)}
	}
	return runToolVersionSpec(ctx, runner, timeout, logPath, tool, executables[0])
}

func runToolVersions(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, tool AppProfile) []CommandResult {
	executables := profileExecutables(tool)
	results := make([]CommandResult, 0, len(executables))
	for _, executable := range executables {
		results = append(results, runToolVersionSpec(ctx, runner, timeout, logPath, tool, executable))
	}
	return results
}

func runToolVersionSpec(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, tool AppProfile, executable ExecutableSpec) CommandResult {
	if !tool.UserBin {
		return runUbuntuLogged(ctx, runner, timeout, logPath, executable.Name, executable.VersionArg...)
	}
	args := append([]string{"-ec", `PATH="$HOME/.local/bin:$PATH"; exec "$@"`, "--", executable.Name}, executable.VersionArg...)
	return runUbuntuLogged(ctx, runner, timeout, logPath, "sh", args...)
}

func firstCommandError(results []CommandResult) error {
	for _, result := range results {
		if result.Err != nil {
			return result.Err
		}
	}
	return nil
}

func commandVersions(profile AppProfile, results []CommandResult) string {
	executables := profileExecutables(profile)
	values := make([]string, 0, len(results))
	for index, result := range results {
		if index >= len(executables) {
			break
		}
		output := commandOutput(result)
		if len(executables) == 1 {
			return output
		}
		values = append(values, executables[index].Name+": "+output)
	}
	return strings.Join(values, "\n")
}

func profilePackages(profile AppProfile) []string {
	if len(profile.Packages) > 0 {
		return append([]string(nil), profile.Packages...)
	}
	if profile.Package == "" {
		return nil
	}
	return []string{profile.Package}
}

func profileExecutables(profile AppProfile) []ExecutableSpec {
	if len(profile.RequiredExecutables) > 0 {
		return append([]ExecutableSpec(nil), profile.RequiredExecutables...)
	}
	if profile.Executable == "" {
		return nil
	}
	return []ExecutableSpec{{Name: profile.Executable, VersionArg: append([]string(nil), profile.VersionArg...)}}
}

func containsExecutable(executables []ExecutableSpec, name string) bool {
	for _, executable := range executables {
		if executable.Name == name {
			return true
		}
	}
	return false
}

func installTool(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, tool AppProfile) CommandResult {
	switch tool.InstallKind {
	case "node":
		return runAptLogged(ctx, runner, timeout, logPath, "install", "-y", "nodejs", "npm")
	case "npm":
		if result := runAptLogged(ctx, runner, timeout, logPath, "install", "-y", "npm"); result.Err != nil {
			return result
		}
		return runUbuntuLogged(ctx, runner, timeout, logPath, "env", "NPM_CONFIG_PREFIX=/root/.local", "npm", "install", "--yes", "-g", tool.Package)
	case "pipx":
		if result := runAptLogged(ctx, runner, timeout, logPath, "install", "-y", "pipx"); result.Err != nil {
			return result
		}
		return runUbuntuLogged(ctx, runner, timeout, logPath, "env", "PIPX_BIN_DIR=/usr/local/bin", "pipx", "install", tool.Package)
	case "go":
		return runUbuntuLogged(ctx, runner, timeout, logPath, "env", "CGO_ENABLED=0", "GOBIN=/usr/local/bin", "go", "install", tool.Package)
	case "ttt":
		if result := runAptLogged(ctx, runner, timeout, logPath, "install", "-y", "git", "ripgrep"); result.Err != nil {
			return result
		}
		return runUbuntuLogged(ctx, runner, timeout, logPath, "env", "CGO_ENABLED=0", "GOBIN=/usr/local/bin", "go", "install", tool.Package)
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
		return runAptLogged(ctx, runner, timeout, logPath, append([]string{"install", "-y"}, profilePackages(tool)...)...)
	}
}

func runAptLogged(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, args ...string) CommandResult {
	if len(args) > 0 && args[0] == "update" {
		args = append([]string{"-y"}, args...)
	}
	aptArgs := append([]string{"-o", fmt.Sprintf("DPkg::Lock::Timeout=%d", aptLockTimeoutSeconds)}, args...)
	return runUbuntuLogged(ctx, runner, timeout, logPath, "apt-get", aptArgs...)
}

func repairDpkg(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string) CommandResult {
	return runUbuntuLogged(ctx, runner, timeout, logPath, "dpkg", "--configure", "-a")
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
	record.LastErrorCode = i18n.ErrorCode(installErr)
	if err := saveRecord(directory, record); err != nil {
		return result, fmt.Errorf("%v; record installation failure: %w", installErr, err)
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
