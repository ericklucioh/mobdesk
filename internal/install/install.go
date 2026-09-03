package install

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
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
	storageWarningBytes   = 20 * 1024 * 1024 * 1024
	storageBlockBytes     = 10 * 1024 * 1024 * 1024
	StorageWarningBytes   = storageWarningBytes
	StorageBlockBytes     = storageBlockBytes
)

// Options configures an installation operation and its external boundaries.
type Options struct {
	Paths          paths.Paths
	Runner         CommandRunner
	Interactive    bool
	Now            func() time.Time
	CommandTimeout time.Duration
	LockTimeout    time.Duration
	Progress       func(string)
	Localizer      i18n.Localizer
	StorageFree    func(string) (int64, error)
}

// Install ensures a named profile and its prerequisites are available.
func Install(ctx context.Context, name string, options Options) (Result, error) {
	options = installDefaults(options)
	release, err := acquireInstallLock(ctx, options)
	if err != nil {
		return Result{}, err
	}
	defer release()
	return install(ctx, name, options)
}

func install(ctx context.Context, name string, options Options) (Result, error) {
	profile, ok := Resolve(name)
	if !ok {
		return Result{}, i18n.NewError(i18n.ServiceInstallUnsupported, "install_unsupported", map[string]any{"Name": name}, nil)
	}
	free, err := availableStorage(options)
	if err != nil {
		return Result{SchemaVersion: 1, Language: profile.Name, State: "failed"}, i18n.NewError(i18n.ServiceInstallStorage, "install_storage_check", map[string]any{"Detail": err.Error()}, err)
	}
	result := Result{SchemaVersion: 1, Language: profile.Name, Package: profile.Package, Packages: profilePackages(profile), Executable: profile.Executable, Executables: profileExecutables(profile), State: "installing", Source: "mobdesk", StorageEstimate: profile.StorageEstimate, StorageFreeBytes: free, StorageWarning: free < storageWarningBytes}
	if free < storageBlockBytes {
		result.State, result.StorageBlocked = "blocked", true
		return result, i18n.NewError(i18n.ServiceInstallStorage, "install_storage_blocked", map[string]any{"Name": profile.Name, "Free": free}, nil)
	}
	for _, prerequisite := range profile.Requires {
		progress(options, i18n.ServiceInstallDependency, map[string]any{"Dependency": prerequisite, "Name": profile.Name})
		if _, err := install(ctx, prerequisite, options); err != nil {
			return result, i18n.NewError(i18n.ServiceInstallDependency, "install_dependency", map[string]any{"Dependency": prerequisite, "Name": profile.Name}, err)
		}
	}
	if err := os.MkdirAll(options.Paths.InstallationsDir(), 0o700); err != nil {
		return result, i18n.NewError(i18n.ServiceInstallState, "install_state", nil, err)
	}
	if err := os.MkdirAll(options.Paths.InstallLogsDir(), 0o700); err != nil {
		return result, i18n.NewError(i18n.ServiceInstallLogs, "install_logs", nil, err)
	}
	logPath := filepath.Join(options.Paths.InstallLogsDir(), profile.Name+".log")
	result.LogPath = logPath
	record := InstallationRecord{Name: profile.Name, Kind: profile.Kind, Package: profile.Package, Packages: profilePackages(profile), Executable: profile.Executable, RequiredExecutables: profileExecutables(profile), Strategy: profile.InstallKind, Dependencies: append([]string(nil), profile.Requires...), InstalledPackages: profilePackages(profile), State: "installing", Source: "mobdesk", LastAttemptAt: options.Now().UTC(), LogPath: logPath}
	if profile.UserBin {
		link, _, directory := managedExecutablePaths(options.Paths, profile)
		record.InstalledFiles = []string{link}
		record.InstalledDirs = []string{directory}
	}
	if err := saveRecord(options.Paths.InstallationsDir(), record); err != nil {
		return result, i18n.NewError(i18n.ServiceInstallRecord, "install_record", nil, err)
	}
	runner := runnerFor(options)
	progress(options, i18n.ServiceInstallVerify, map[string]any{"Name": profile.Name})
	versions := runToolVersions(ctx, runner, options.CommandTimeout, logPath, options.Paths, profile)
	if firstCommandError(versions) != nil {
		progress(options, i18n.ServiceInstallTool, map[string]any{"Name": profile.Name})
		if installed := installTool(ctx, runner, options.CommandTimeout, logPath, options.Paths, profile); installed.Err != nil {
			return failInstallation(options.Paths.InstallationsDir(), record, result, i18n.NewError(i18n.ServiceInstallTool, "install_tool", map[string]any{"Name": profile.Name}, installed.Err))
		}
		result.Changed = true
		versions = runToolVersions(ctx, runner, options.CommandTimeout, logPath, options.Paths, profile)
	}
	if verifyErr := firstCommandError(versions); verifyErr != nil {
		return failInstallation(options.Paths.InstallationsDir(), record, result, i18n.NewError(i18n.ServiceInstallVerify, "install_verify", map[string]any{"Name": profile.Name}, verifyErr))
	}
	result.Version, result.Installed, result.State = commandVersions(profile, versions), true, "installed"
	result.Paths = append(result.Paths, record.InstalledFiles...)
	record.State, record.Version, record.InstalledAt = result.State, result.Version, options.Now().UTC()
	if profile.Name == "java" {
		javaHome, discoverErr := discoverJavaHome(ctx, runner, options.CommandTimeout, logPath, options.Paths.Prefix)
		if discoverErr != nil {
			return failInstallation(options.Paths.InstallationsDir(), record, result, i18n.NewError(i18n.ServiceInstallVerify, "install_java_home", map[string]any{"Name": profile.Name}, discoverErr))
		}
		result.JavaHome, record.JavaHome = javaHome, javaHome
	}
	if profile.UserBin {
		_, target, _ := managedExecutablePaths(options.Paths, profile)
		digest, hashErr := fileSHA256(target)
		if hashErr != nil {
			return failInstallation(options.Paths.InstallationsDir(), record, result, i18n.NewError(i18n.ServiceInstallVerify, "install_managed_file", map[string]any{"Name": profile.Name}, hashErr))
		}
		record.InstalledFileHashes = map[string]string{target: digest}
	}
	if err := saveRecord(options.Paths.InstallationsDir(), record); err != nil {
		return result, i18n.NewError(i18n.ServiceInstallRecord, "install_record", nil, err)
	}
	return result, nil
}

func discoverJavaHome(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath, prefix string) (string, error) {
	result := runTermuxLogged(ctx, runner, timeout, logPath, "java", "-XshowSettings:properties", "-version")
	if result.Err != nil {
		return "", fmt.Errorf("read java.home: %w", result.Err)
	}
	return parseJavaHome(string(result.Stdout)+"\n"+string(result.Stderr), prefix)
}

func parseJavaHome(output, prefix string) (string, error) {
	prefix = filepath.Clean(strings.TrimSpace(prefix))
	if !filepath.IsAbs(prefix) {
		return "", fmt.Errorf("Termux prefix is not absolute")
	}
	for _, line := range strings.Split(output, "\n") {
		name, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || strings.TrimSpace(name) != "java.home" {
			continue
		}
		home := filepath.Clean(strings.TrimSpace(value))
		relative, err := filepath.Rel(prefix, home)
		if !filepath.IsAbs(home) || err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("java.home %q is outside Termux prefix %q", home, prefix)
		}
		return home, nil
	}
	return "", fmt.Errorf("java.home was not reported by the runtime")
}

func installDefaults(options Options) Options {
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.CommandTimeout <= 0 {
		options.CommandTimeout = defaultCommandTimeout
	}
	if options.LockTimeout <= 0 {
		options.LockTimeout = defaultLockTimeout
	}
	if options.Localizer.Locale == "" {
		options.Localizer = i18n.New(i18n.LocaleENUS)
	}
	return options
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
	if options.Progress == nil {
		return
	}
	if data == nil {
		data = map[string]any{}
	}
	if _, ok := data["Detail"]; !ok {
		data["Detail"] = ""
	}
	options.Progress(options.Localizer.Text(id, data))
}

func runToolVersions(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, p paths.Paths, profile AppProfile) []CommandResult {
	results := make([]CommandResult, 0, len(profileExecutables(profile)))
	for _, executable := range profileExecutables(profile) {
		results = append(results, runTermuxLogged(ctx, runner, timeout, logPath, executablePath(p, profile, executable.Name), executable.VersionArg...))
	}
	return results
}

func installTool(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, p paths.Paths, profile AppProfile) CommandResult {
	switch profile.InstallKind {
	case "pkg":
		return runTermuxLogged(ctx, runner, timeout, logPath, "pkg", append([]string{"install", "-y"}, profilePackages(profile)...)...)
	case "pipx":
		return installPipx(ctx, runner, timeout, logPath, p, profile)
	case "npm":
		return installNPM(ctx, runner, timeout, logPath, p, profile)
	case "go":
		return installGo(ctx, runner, timeout, logPath, p, profile)
	default:
		return CommandResult{Err: fmt.Errorf("unsupported native install strategy %q", profile.InstallKind)}
	}
}

func installPipx(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, p paths.Paths, profile AppProfile) CommandResult {
	link, target, directory := managedExecutablePaths(p, profile)
	if err := ensureManagedLink(link, target); err != nil {
		return CommandResult{Err: err}
	}
	runtime := filepath.Join(p.ManagedToolsDir(), "pipx", "runtime")
	pipx := filepath.Join(runtime, "bin", "pipx")
	if _, err := os.Stat(pipx); os.IsNotExist(err) {
		if result := runTermuxLogged(ctx, runner, timeout, logPath, "python", "-m", "venv", runtime); result.Err != nil {
			return result
		}
		if result := runTermuxLogged(ctx, runner, timeout, logPath, filepath.Join(runtime, "bin", "python"), "-m", "pip", "install", "--disable-pip-version-check", "pipx"); result.Err != nil {
			return result
		}
	} else if err != nil {
		return CommandResult{Err: err}
	}
	home := filepath.Join(directory, "home")
	bin := filepath.Dir(target)
	if err := os.MkdirAll(bin, 0o700); err != nil {
		return CommandResult{Err: err}
	}
	result := runWithEnvironment(ctx, runner, timeout, logPath, []string{"ANDROID_API_LEVEL=24", "PIPX_HOME=" + home, "PIPX_BIN_DIR=" + bin, "PIPX_DEFAULT_PYTHON=python"}, pipx, "install", "--force", profile.Package)
	if result.Err != nil {
		return result
	}
	if err := publishManagedLink(link, target); err != nil {
		return CommandResult{Err: err}
	}
	return result
}

func installNPM(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, p paths.Paths, profile AppProfile) CommandResult {
	link, target, directory := managedExecutablePaths(p, profile)
	if err := ensureManagedLink(link, target); err != nil {
		return CommandResult{Err: err}
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return CommandResult{Err: err}
	}
	args := []string{"install", "--global", "--prefix", directory, "--cache", filepath.Join(directory, "cache"), "--no-audit", "--no-fund"}
	if profile.Name == "pi" {
		args = append(args, "--ignore-scripts")
	}
	args = append(args, profile.Package)
	result := runTermuxLogged(ctx, runner, timeout, logPath, "npm", args...)
	if result.Err != nil {
		return result
	}
	if err := writeNPMLauncher(p, profile, target, directory); err != nil {
		return CommandResult{Err: err}
	}
	if err := publishManagedLink(link, target); err != nil {
		return CommandResult{Err: err}
	}
	return result
}

func writeNPMLauncher(p paths.Paths, profile AppProfile, target, directory string) error {
	var entrypoint string
	switch {
	case profile.Name == "bitwarden" && profile.Package == "@bitwarden/cli@2025.12.0":
		entrypoint = filepath.Join(directory, "lib", "node_modules", "@bitwarden", "cli", "build", "bw.js")
	case profile.Name == "pi" && profile.Package == "@earendil-works/pi-coding-agent@0.84.4":
		entrypoint = filepath.Join(directory, "lib", "node_modules", "@earendil-works", "pi-coding-agent", "dist", "bundle", "cli.js")
	default:
		return fmt.Errorf("unsupported managed npm profile %q", profile.Name)
	}
	if _, err := os.Stat(entrypoint); err != nil {
		return fmt.Errorf("managed npm entrypoint %q was not created: %w", entrypoint, err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	content := "#!" + filepath.Join(p.Prefix, "bin", "sh") + "\nexec " + shellQuote(filepath.Join(p.Prefix, "bin", "node")) + " " + shellQuote(entrypoint) + " \"$@\"\n"
	return os.WriteFile(target, []byte(content), 0o700)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func installGo(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, p paths.Paths, profile AppProfile) CommandResult {
	link, target, directory := managedExecutablePaths(p, profile)
	if err := ensureManagedLink(link, target); err != nil {
		return CommandResult{Err: err}
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return CommandResult{Err: err}
	}
	result := runWithEnvironment(ctx, runner, timeout, logPath, []string{
		"GOBIN=" + filepath.Dir(target),
		"GOPATH=" + filepath.Join(directory, "gopath"),
		"GOCACHE=" + filepath.Join(directory, "cache"),
		"GOFLAGS=-modcacherw",
	}, "go", "install", profile.Package)
	if result.Err != nil {
		return result
	}
	if err := publishManagedLink(link, target); err != nil {
		return CommandResult{Err: err}
	}
	return result
}

func runWithEnvironment(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, environment []string, name string, args ...string) CommandResult {
	command := append(append([]string(nil), environment...), name)
	command = append(command, args...)
	return runTermuxLogged(ctx, runner, timeout, logPath, "env", command...)
}

func executablePath(p paths.Paths, profile AppProfile, executable string) string {
	if !profile.UserBin {
		return executable
	}
	return filepath.Join(p.UserBinDir(), executable)
}

func managedExecutablePaths(p paths.Paths, profile AppProfile) (link, target, directory string) {
	link = filepath.Join(p.UserBinDir(), profile.Executable)
	switch profile.InstallKind {
	case "pipx":
		directory = filepath.Join(p.ManagedToolsDir(), "pipx", profile.Name)
	case "npm":
		directory = filepath.Join(p.ManagedToolsDir(), "npm", profile.Name)
		target = filepath.Join(directory, "launcher", profile.Executable)
	case "go":
		directory = filepath.Join(p.ManagedToolsDir(), "go", profile.Name)
	}
	if target == "" {
		target = filepath.Join(directory, "bin", profile.Executable)
	}
	return link, target, directory
}

func ensureManagedLink(link, target string) error {
	info, err := os.Lstat(link)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("managed executable path %q already exists", link)
	}
	current, err := os.Readlink(link)
	if err != nil {
		return err
	}
	if current != target {
		return fmt.Errorf("managed executable path %q points outside Mobdesk", link)
	}
	return nil
}

func publishManagedLink(link, target string) error {
	if _, err := os.Stat(target); err != nil {
		return fmt.Errorf("managed executable %q was not created: %w", target, err)
	}
	if err := ensureManagedLink(link, target); err != nil {
		return err
	}
	if _, err := os.Lstat(link); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(link), 0o700); err != nil {
		return err
	}
	return os.Symlink(target, link)
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func runTermuxLogged(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath, name string, args ...string) CommandResult {
	commandContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	result := runner.Run(commandContext, name, args...)
	_ = appendLog(logPath, name, args, result)
	return result
}

func appendLog(path, name string, args []string, result CommandResult) error {
	if path == "" {
		return nil
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if _, err := fmt.Fprintf(file, "\n$ %s %s\n", name, strings.Join(args, " ")); err != nil {
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
	for {
		if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err == nil {
			return func() { _ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN); _ = file.Close() }, nil
		}
		if err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
			_ = file.Close()
			return nil, i18n.NewError(i18n.ServiceInstallLock, "install_lock", nil, err)
		}
		progress(options, i18n.ServiceInstallWait, nil)
		select {
		case <-ctx.Done():
			_ = file.Close()
			return nil, i18n.NewError(i18n.ServiceInstallWait, "install_wait", nil, ctx.Err())
		case <-time.After(250 * time.Millisecond):
		}
	}
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
func firstCommandError(results []CommandResult) error {
	for _, result := range results {
		if result.Err != nil {
			return result.Err
		}
	}
	return nil
}
func commandVersions(profile AppProfile, results []CommandResult) string {
	values := make([]string, 0, len(results))
	executables := profileExecutables(profile)
	for index, result := range results {
		output := strings.TrimSpace(string(result.Stdout))
		if output == "" {
			output = strings.TrimSpace(string(result.Stderr))
		}
		if len(executables) == 1 {
			return output
		}
		values = append(values, executables[index].Name+": "+output)
	}
	return strings.Join(values, "\n")
}

func saveRecord(directory string, record InstallationRecord) error {
	payload, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, record.Name+"-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
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
	return os.Rename(temporaryPath, filepath.Join(directory, record.Name+".json"))
}

func failInstallation(directory string, record InstallationRecord, result Result, installErr error) (Result, error) {
	result.State, record.State, record.LastError, record.LastErrorCode = "failed", "failed", installErr.Error(), i18n.ErrorCode(installErr)
	if err := saveRecord(directory, record); err != nil {
		return result, fmt.Errorf("%v; record installation failure: %w", installErr, err)
	}
	return result, installErr
}
