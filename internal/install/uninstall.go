package install

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	if packageSharedByAnotherInstallation(options.Paths, record) {
		return result, i18n.NewError(i18n.ServiceUninstallShared, "uninstall_shared_package", map[string]any{"Name": profile.Name}, nil)
	}
	record.State, record.LastAttemptAt = "uninstalling", options.Now().UTC()
	if err := saveRecord(options.Paths.InstallationsDir(), record); err != nil {
		return result, i18n.NewError(i18n.ServiceUninstallError, "uninstall_record", nil, err)
	}
	progress(options, i18n.ServiceUninstallProgress, map[string]any{"Name": profile.Name})
	if record.Strategy != "pkg" {
		return result, fmt.Errorf("unsupported native uninstall strategy %q", record.Strategy)
	}
	removed := runTermuxLogged(ctx, runnerFor(options), options.CommandTimeout, record.LogPath, "pkg", append([]string{"uninstall", "-y"}, record.Packages...)...)
	if removed.Err != nil {
		record.State, record.LastError, record.LastErrorCode = "failed", removed.Err.Error(), i18n.ErrorCode(removed.Err)
		_ = saveRecord(options.Paths.InstallationsDir(), record)
		return result, removed.Err
	}
	record.State, record.LastError, record.RemovedPackages = "uninstalled", "", append(record.RemovedPackages, record.Packages...)
	if err := saveRecord(options.Paths.InstallationsDir(), record); err != nil {
		return result, i18n.NewError(i18n.ServiceUninstallError, "uninstall_record", nil, err)
	}
	result.State, result.Changed = "uninstalled", true
	return result, nil
}

func packageSharedByAnotherInstallation(p paths.Paths, target InstallationRecord) bool {
	entries, err := os.ReadDir(p.InstallationsDir())
	if err != nil {
		return false
	}
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
					return true
				}
			}
		}
	}
	return false
}
