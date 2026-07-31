package install

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/paths"
)

func Uninstall(ctx context.Context, name string, options Options) (result Result, err error) {
	defer func() {
		if err != nil && i18n.ErrorCode(err) == "" {
			err = i18n.NewError(i18n.ServiceUninstallError, "uninstall_operation_failed", map[string]any{"Detail": err.Error()}, err)
		}
	}()
	if options.Localizer.Locale == "" {
		options.Localizer = i18n.New(i18n.LocaleENUS)
	}
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

	profile, ok := Resolve(name)
	if !ok {
		return Result{}, i18n.NewError(i18n.ServiceUninstallError, "uninstall_unsupported", map[string]any{"Detail": name}, nil)
	}
	result = Result{
		SchemaVersion:   1,
		Language:        profile.Name,
		Package:         profile.Package,
		Executable:      profile.Executable,
		State:           "failed",
		Source:          "mobdesk",
		StorageEstimate: profile.StorageEstimate,
	}
	record, err := loadInstallationRecord(options.Paths, profile.Name)
	if err != nil {
		return result, i18n.NewError(i18n.ServiceUninstallError, "uninstall_load", map[string]any{"Detail": profile.Name}, err)
	}
	if record.Source == "" {
		record.Source = "mobdesk"
	}
	result.Source = record.Source
	if record.Source != "mobdesk" {
		return result, i18n.NewError(i18n.ServiceUninstallDetected, "uninstall_detected", map[string]any{"Name": profile.Name}, nil)
	}
	if record.State == "uninstalled" {
		result.State = "uninstalled"
		result.Installed = false
		return result, nil
	}
	if record.State != "installed" && record.State != "partial" && record.State != "failed" {
		return result, i18n.NewError(i18n.ServiceUninstallState, "uninstall_invalid_state", map[string]any{"Name": profile.Name, "State": record.State}, nil)
	}
	if packageSharedByAnotherInstallation(options.Paths, record) {
		return result, i18n.NewError(i18n.ServiceUninstallShared, "uninstall_shared_package", map[string]any{"Name": profile.Name}, nil)
	}

	record.State = "uninstalling"
	record.LastAttemptAt = options.Now().UTC()
	if err := saveRecord(options.Paths.InstallationsDir(), record); err != nil {
		return result, i18n.NewError(i18n.ServiceUninstallError, "uninstall_record", nil, err)
	}

	strategy := record.Strategy
	if strategy == "" {
		strategy = profile.InstallKind
	}
	progress(options, i18n.ServiceUninstallProgress, map[string]any{"Name": profile.Name, "Detail": ""})
	runner := runnerFor(options)
	removedFiles, preservedFiles, removeErr := uninstallStrategy(ctx, runner, options, strategy, record)
	result.Paths = append(removedFiles, preservedFiles...)
	result.Conflicts = append([]string(nil), preservedFiles...)
	if removeErr != nil {
		record.State = "failed"
		record.LastError = removeErr.Error()
		record.LastErrorCode = i18n.ErrorCode(removeErr)
		record.RemovedFiles = append(record.RemovedFiles, removedFiles...)
		record.PreservedFiles = append(record.PreservedFiles, preservedFiles...)
		_ = saveRecord(options.Paths.InstallationsDir(), record)
		return result, removeErr
	}

	record.State = "uninstalled"
	record.LastError = ""
	record.RemovedFiles = append(record.RemovedFiles, removedFiles...)
	record.PreservedFiles = append(record.PreservedFiles, preservedFiles...)
	if record.Package != "" && strategy != "script" && strategy != "go" && strategy != "ttt" && strategy != "cargo" && strategy != "gh-extension" {
		record.RemovedPackages = append(record.RemovedPackages, record.Package)
	}
	if len(record.PreservedFiles) > 0 {
		record.State = "modified"
	}
	progress(options, i18n.ServiceUninstallProgress, map[string]any{"Name": profile.Name, "Detail": ""})
	if err := saveRecord(options.Paths.InstallationsDir(), record); err != nil {
		return result, i18n.NewError(i18n.ServiceUninstallError, "uninstall_record", nil, err)
	}
	result.State = record.State
	result.Installed = false
	result.Changed = true
	return result, nil
}

func uninstallStrategy(ctx context.Context, runner CommandRunner, options Options, strategy string, record InstallationRecord) ([]string, []string, error) {
	switch strategy {
	case "apt", "node":
		if record.Package == "" {
			return nil, nil, fmt.Errorf("package missing from %s record", record.Name)
		}
		result := runAptLogged(ctx, runner, options.CommandTimeout, record.LogPath, "remove", "-y", record.Package)
		return nil, nil, result.Err
	case "npm":
		if record.Package == "" {
			return nil, nil, fmt.Errorf("package missing from %s record", record.Name)
		}
		result := runUbuntuLogged(ctx, runner, options.CommandTimeout, record.LogPath, "env", "NPM_CONFIG_PREFIX=/root/.local", "npm", "uninstall", "-g", record.Package)
		return nil, nil, result.Err
	case "pipx":
		if record.Package == "" {
			return nil, nil, fmt.Errorf("package missing from %s record", record.Name)
		}
		result := runUbuntuLogged(ctx, runner, options.CommandTimeout, record.LogPath, "pipx", "uninstall", record.Package)
		return nil, nil, result.Err
	case "script", "go", "ttt", "cargo", "gh-extension":
		if len(record.InstalledFiles) == 0 {
			return nil, nil, fmt.Errorf("no verified files to remove for %s", record.Name)
		}
		return removeTrackedFiles(ctx, runner, options, record)
	default:
		return nil, nil, fmt.Errorf("unsupported uninstall strategy %q", strategy)
	}
}

func removeTrackedFiles(ctx context.Context, runner CommandRunner, options Options, record InstallationRecord) ([]string, []string, error) {
	removed := []string{}
	preserved := []string{}
	for _, path := range record.InstalledFiles {
		if err := validateManagedPath(path); err != nil {
			return removed, preserved, err
		}
		exists := runUbuntuLogged(ctx, runner, options.CommandTimeout, record.LogPath, "test", "-e", path)
		if exists.Err != nil {
			continue
		}
		expected, ok := record.InstalledFileHashes[path]
		if !ok || expected == "" {
			preserved = append(preserved, path)
			continue
		}
		current, err := currentFileHash(ctx, runner, options, record.LogPath, path)
		if err != nil || current != expected {
			preserved = append(preserved, path)
			continue
		}
		removedResult := runUbuntuLogged(ctx, runner, options.CommandTimeout, record.LogPath, "rm", "--", path)
		if removedResult.Err != nil {
			return removed, preserved, removedResult.Err
		}
		removed = append(removed, path)
	}
	return removed, preserved, nil
}

func currentFileHash(ctx context.Context, runner CommandRunner, options Options, logPath, path string) (string, error) {
	result := runUbuntuLogged(ctx, runner, options.CommandTimeout, logPath, "sha256sum", "--", path)
	if result.Err != nil {
		return "", result.Err
	}
	fields := strings.Fields(string(result.Stdout))
	if len(fields) == 0 {
		return "", fmt.Errorf("empty hash for %s", path)
	}
	return fields[0], nil
}

func validateManagedPath(path string) error {
	clean := filepath.Clean(path)
	if clean != path || (!strings.HasPrefix(path, "/root/.local/bin/") && !strings.HasPrefix(path, "/usr/local/bin/")) {
		return fmt.Errorf("invalid managed file %q", path)
	}
	return nil
}

func packageSharedByAnotherInstallation(p paths.Paths, target InstallationRecord) bool {
	if target.Package == "" {
		return false
	}
	entries, err := os.ReadDir(p.InstallationsDir())
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" || entry.Name() == target.Name+".json" {
			continue
		}
		payload, err := os.ReadFile(filepath.Join(p.InstallationsDir(), entry.Name()))
		if err != nil {
			continue
		}
		var record InstallationRecord
		if json.Unmarshal(payload, &record) != nil || record.Package != target.Package {
			continue
		}
		if record.Source == "" {
			record.Source = "mobdesk"
		}
		if record.Source == "mobdesk" && (record.State == "installed" || record.State == "installing" || record.State == "uninstalling") {
			return true
		}
	}
	return false
}
