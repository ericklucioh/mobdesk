package install

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultCommandTimeout = 10 * time.Minute

var catalog = []Language{
	{Name: "go", Aliases: []string{"golang"}, Package: "golang", Executable: "go", VersionArg: []string{"version"}},
	{Name: "python", Aliases: []string{"python3"}, Package: "python3", Executable: "python3", VersionArg: []string{"--version"}},
	{Name: "node", Aliases: []string{"nodejs"}, Package: "nodejs", Executable: "node", VersionArg: []string{"--version"}},
	{Name: "c", Aliases: []string{"c-lang"}, Package: "clang", Executable: "clang", VersionArg: []string{"--version"}},
	{Name: "cpp", Aliases: []string{"c++", "cplusplus"}, Package: "clang", Executable: "clang++", VersionArg: []string{"--version"}},
	{Name: "lua", Aliases: []string{"lua5.4"}, Package: "lua5.4", Executable: "lua5.4", VersionArg: []string{"-v"}},
}

type Options struct {
	Runner           CommandRunner
	InstallationsDir string
	LogsDir          string
	Now              func() time.Time
	CommandTimeout   time.Duration
}

func Languages() []Language {
	result := make([]Language, len(catalog))
	copy(result, catalog)
	return result
}

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
	language, ok := Resolve(name)
	if !ok {
		return Result{}, fmt.Errorf("linguagem não suportada %q", name)
	}
	runner := options.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.CommandTimeout <= 0 {
		options.CommandTimeout = defaultCommandTimeout
	}
	if options.InstallationsDir == "" || options.LogsDir == "" {
		home := os.Getenv("HOME")
		if home == "" {
			home = "."
		}
		base := filepath.Join(home, ".local", "share", "mobdesk")
		if options.InstallationsDir == "" {
			options.InstallationsDir = filepath.Join(base, "state", "installations")
		}
		if options.LogsDir == "" {
			options.LogsDir = filepath.Join(base, "logs", "install")
		}
	}

	now := options.Now().UTC()
	logPath := filepath.Join(options.LogsDir, language.Name+".log")
	result := Result{
		Language:   language.Name,
		Package:    language.Package,
		Executable: language.Executable,
		State:      "installing",
		LogPath:    logPath,
	}
	record := InstallationRecord{
		Name:          language.Name,
		Kind:          "language",
		Package:       language.Package,
		Executable:    language.Executable,
		State:         "installing",
		LastAttemptAt: now,
		LogPath:       logPath,
	}
	if err := os.MkdirAll(options.InstallationsDir, 0o700); err != nil {
		return result, fmt.Errorf("criar estado da instalação: %w", err)
	}
	if err := os.MkdirAll(options.LogsDir, 0o700); err != nil {
		return result, fmt.Errorf("criar diretório de logs da instalação: %w", err)
	}
	if err := saveRecord(options.InstallationsDir, record); err != nil {
		return result, fmt.Errorf("registrar tentativa de instalação: %w", err)
	}

	version := runUbuntuLogged(ctx, runner, options.CommandTimeout, logPath, language.Executable, language.VersionArg...)
	if version.Err != nil {
		if update := runUbuntuLogged(ctx, runner, options.CommandTimeout, logPath, "apt-get", "update"); update.Err != nil {
			err := fmt.Errorf("atualizar índices do Ubuntu para %s: %w", language.Name, update.Err)
			return failInstallation(options.InstallationsDir, record, result, err)
		}
		if install := runUbuntuLogged(ctx, runner, options.CommandTimeout, logPath, "apt-get", "install", "-y", language.Package); install.Err != nil {
			err := fmt.Errorf("instalar %s: %w", language.Name, install.Err)
			return failInstallation(options.InstallationsDir, record, result, err)
		}
		result.Changed = true
		version = runUbuntuLogged(ctx, runner, options.CommandTimeout, logPath, language.Executable, language.VersionArg...)
	}
	if version.Err != nil {
		err := fmt.Errorf("verificar %s após instalação: %w", language.Name, version.Err)
		return failInstallation(options.InstallationsDir, record, result, err)
	}
	result.Version = commandOutput(version)
	result.Installed = true
	result.State = "installed"
	record.State = result.State
	record.Version = result.Version
	record.InstalledAt = options.Now().UTC()
	if err := saveRecord(options.InstallationsDir, record); err != nil {
		return result, fmt.Errorf("registrar instalação concluída: %w", err)
	}
	return result, nil
}

func runUbuntu(ctx context.Context, runner CommandRunner, name string, args ...string) CommandResult {
	loginArgs := append([]string{"login", "ubuntu", "--", name}, args...)
	return runner.Run(ctx, "proot-distro", loginArgs...)
}

func runUbuntuLogged(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, name string, args ...string) CommandResult {
	commandContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	loginArgs := append([]string{"login", "ubuntu", "--", name}, args...)
	result := runner.Run(commandContext, "proot-distro", loginArgs...)
	_ = appendLog(logPath, loginArgs, result)
	return result
}

func appendLog(path string, args []string, result CommandResult) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
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
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(append(payload, '\n')); err != nil {
		temporary.Close()
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
