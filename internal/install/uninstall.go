package install

import (
	"context"
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
	options = installDefaults(options)
	release, err := acquireInstallLock(ctx, options)
	if err != nil {
		return Result{}, err
	}
	defer release()
	profile, ok := Resolve(name)
	if !ok {
		return Result{}, i18n.NewError(i18n.ServiceUninstallError, "uninstall_unsupported", map[string]any{"Detail": name}, nil)
	}
	result = Result{SchemaVersion: 1, Language: profile.Name, Package: profile.Package, Packages: profilePackages(profile), Executable: profile.Executable, Executables: profileExecutables(profile), State: "failed", Source: "mobdesk", StorageEstimate: profile.StorageEstimate}
	record, err := loadInstallationRecord(options.Paths, profile.Name)
	if err != nil {
		return result, i18n.NewError(i18n.ServiceUninstallError, "uninstall_load", map[string]any{"Detail": profile.Name}, err)
	}
	if record.Source == "" {
		record.Source = "mobdesk"
	}
	if record.Source != "mobdesk" {
		return result, i18n.NewError(i18n.ServiceUninstallDetected, "uninstall_detected", map[string]any{"Name": profile.Name}, nil)
	}
	if record.State == "uninstalled" {
		result.State = "uninstalled"
		return result, nil
	}
	if record.State != "installed" && record.State != "failed" && record.State != "partial" {
		return result, i18n.NewError(i18n.ServiceUninstallState, "uninstall_invalid_state", map[string]any{"Name": profile.Name, "State": record.State}, nil)
	}
	if dependents := dependentInstallations(options.Paths, record.Name); len(dependents) > 0 {
		result.Conflicts = dependents
		return result, i18n.NewError(i18n.ServiceUninstallRequired, "uninstall_required", map[string]any{"Name": profile.Name, "Dependents": strings.Join(dependents, ", ")}, nil)
	}
	if profile.UserBin {
		link, target, _ := managedExecutablePaths(options.Paths, profile)
		if err := ensureManagedLink(link, target); err != nil {
			result.Conflicts = []string{link}
			return result, err
		}
		for path, digest := range record.InstalledFileHashes {
			current, err := fileSHA256(path)
			if err != nil || current != digest {
				result.Conflicts = []string{path}
				return result, fmt.Errorf("managed file %q was changed", path)
			}
		}
	}
	sharedPackages := []string(nil)
	if record.Strategy == "pkg" {
		sharedPackages = sharedPackagesByAnotherInstallation(options.Paths, record)
	}
	record.State, record.LastAttemptAt = "uninstalling", options.Now().UTC()
	if err := saveRecord(options.Paths.InstallationsDir(), record); err != nil {
		return result, i18n.NewError(i18n.ServiceUninstallError, "uninstall_record", nil, err)
	}
	progress(options, i18n.ServiceUninstallProgress, map[string]any{"Name": profile.Name})
	if len(sharedPackages) > 0 {
		record.State, record.LastError = "uninstalled", ""
		if err := saveRecord(options.Paths.InstallationsDir(), record); err != nil {
			return result, i18n.NewError(i18n.ServiceUninstallError, "uninstall_record", nil, err)
		}
		result.State, result.Changed, result.PreservedPackages = "uninstalled", true, sharedPackages
		return result, nil
	}
	removed := uninstallTool(ctx, runnerFor(options), options.CommandTimeout, record.LogPath, options.Paths, profile, record.Strategy)
	if removed.Err != nil {
		record.State, record.LastError, record.LastErrorCode = "failed", removed.Err.Error(), i18n.ErrorCode(removed.Err)
		_ = saveRecord(options.Paths.InstallationsDir(), record)
		return result, removed.Err
	}
	record.State, record.LastError = "uninstalled", ""
	if record.Strategy == "pkg" {
		record.RemovedPackages = append(record.RemovedPackages, record.Packages...)
	} else {
		record.RemovedFiles = append(record.RemovedFiles, record.InstalledFiles...)
		result.Paths = append(result.Paths, record.InstalledFiles...)
		for _, directory := range record.InstalledDirs {
			if err := os.RemoveAll(directory); err != nil {
				record.PreservedFiles = append(record.PreservedFiles, directory)
				result.Conflicts = append(result.Conflicts, directory)
			}
		}
	}
	if err := saveRecord(options.Paths.InstallationsDir(), record); err != nil {
		return result, i18n.NewError(i18n.ServiceUninstallError, "uninstall_record", nil, err)
	}
	result.State, result.Changed = "uninstalled", true
	return result, nil
}

func uninstallTool(ctx context.Context, runner CommandRunner, timeout time.Duration, logPath string, p paths.Paths, profile AppProfile, strategy string) CommandResult {
	switch strategy {
	case "pkg":
		return runTermuxLogged(ctx, runner, timeout, logPath, "pkg", append([]string{"uninstall", "-y"}, profilePackages(profile)...)...)
	case "pipx":
		link, target, directory := managedExecutablePaths(p, profile)
		home := filepath.Join(directory, "home")
		bin := filepath.Dir(target)
		pipx := filepath.Join(p.ManagedToolsDir(), "pipx", "runtime", "bin", "pipx")
		result := runWithEnvironment(ctx, runner, timeout, logPath, []string{"PIPX_HOME=" + home, "PIPX_BIN_DIR=" + bin, "PIPX_DEFAULT_PYTHON=python"}, pipx, "uninstall", pipxPackageName(profile.Package))
		if result.Err == nil {
			result.Err = removeManagedLink(link, target)
		}
		return result
	default:
		return CommandResult{Err: fmt.Errorf("unsupported native uninstall strategy %q", strategy)}
	}
}

func pipxPackageName(value string) string {
	name, _, _ := strings.Cut(value, "==")
	return name
}

func removeManagedLink(path, target string) error {
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	if err := ensureManagedLink(path, target); err != nil {
		return err
	}
	return os.Remove(path)
}

func dependentInstallations(p paths.Paths, dependency string) []string {
	entries, err := os.ReadDir(p.InstallationsDir())
	if err != nil {
		return nil
	}
	var result []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		record, err := loadInstallationRecord(p, strings.TrimSuffix(entry.Name(), ".json"))
		if err != nil || record.State != "installed" {
			continue
		}
		for _, current := range record.Dependencies {
			if current == dependency {
				result = append(result, record.Name)
				break
			}
		}
	}
	return result
}

func sharedPackagesByAnotherInstallation(p paths.Paths, target InstallationRecord) []string {
	entries, err := os.ReadDir(p.InstallationsDir())
	if err != nil {
		return nil
	}
	shared := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == target.Name+".json" {
			continue
		}
		record, err := loadInstallationRecord(p, strings.TrimSuffix(entry.Name(), ".json"))
		if err != nil || record.State != "installed" {
			continue
		}
		for _, current := range record.Packages {
			for _, targetPackage := range target.Packages {
				if current == targetPackage {
					shared[targetPackage] = true
				}
			}
		}
	}
	result := make([]string, 0, len(shared))
	for _, targetPackage := range target.Packages {
		if shared[targetPackage] {
			result = append(result, targetPackage)
		}
	}
	return result
}
