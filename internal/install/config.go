package install

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ericklucioh/mobdesk/internal/i18n"
)

const configWriteScript = `umask 077; tmp=$(mktemp "$1.tmp.XXXXXX"); printf '%s' "$2" | base64 -d > "$tmp"; chmod "$3" "$tmp"; mv "$tmp" "$1"`

func ApplyConfig(ctx context.Context, app string, options Options) (ConfigOperationResult, error) {
	options = configDefaults(options)
	return applyConfig(ctx, app, options)
}

func RemoveConfig(ctx context.Context, app string, options Options) (ConfigOperationResult, error) {
	options = configDefaults(options)
	return removeConfig(ctx, app, options)
}

func applyConfig(ctx context.Context, app string, options Options) (ConfigOperationResult, error) {
	profile, appProfile, result, err := resolveConfigProfile(app, options)
	if err != nil {
		return result, err
	}
	result.Action = "apply"
	result.Profile = profile.ID
	installation, err := loadInstallationRecord(options.Paths, appProfile.Name)
	if err != nil {
		result.State = ConfigStateFailed
		return result, i18n.NewError(i18n.ServiceConfigError, "config_load_installation", map[string]any{"Detail": fmt.Sprintf("%s", err)}, err)
	}
	if installation.State != "installed" || installation.Source == "detected" {
		result.State = ConfigStateFailed
		return result, i18n.NewError(i18n.ServiceConfigError, "config_installation_required", map[string]any{"Detail": fmt.Sprintf("%s must be installed by Mobdesk", appProfile.Name)}, nil)
	}
	if err := validateConfigProfile(profile); err != nil {
		result.State = ConfigStateFailed
		return result, err
	}

	existing, existingErr := LoadConfigurationRecord(options.Paths, appProfile.Name)
	if existingErr == nil {
		if existing.Profile != profile.ID || existing.ProfileVersion != profile.Version {
			result.State = ConfigStateConflict
			result.Conflicts = []string{appProfile.ConfigTarget}
			return result, i18n.NewError(i18n.ServiceConfigError, "config_profile_conflict", map[string]any{"Detail": "existing configuration uses another profile"}, nil)
		}
		if existing.State == ConfigStateApplied {
			for path, expected := range existing.FileHashes {
				current, hashErr := currentFileHash(ctx, runnerFor(options), options, "", path)
				if hashErr != nil || current != expected {
					result.State = ConfigStateConflict
					result.Conflicts = append(result.Conflicts, path)
					return result, i18n.NewError(i18n.ServiceConfigError, "config_modified_file", map[string]any{"Detail": path}, nil)
				}
			}
			result.State = ConfigStateApplied
			result.Success = true
			result.Message = options.Localizer.Text(i18n.OutputConfigApplied, nil)
			result.MessageID = string(i18n.OutputConfigApplied)
			return result, nil
		}
	} else if !errors.Is(existingErr, os.ErrNotExist) {
		result.State = ConfigStateFailed
		return result, existingErr
	}

	runner := runnerFor(options)
	paths := configPaths(profile)
	for _, path := range paths {
		if exists := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "test", "-e", path); exists.Err == nil {
			result.State = ConfigStateConflict
			result.Conflicts = append(result.Conflicts, path)
		}
	}
	if len(result.Conflicts) > 0 {
		return result, i18n.NewError(i18n.ServiceConfigError, "config_path_conflict", map[string]any{"Detail": "existing configuration conflicts"}, nil)
	}

	record := ConfigurationRecord{
		App:            appProfile.Name,
		Profile:        profile.ID,
		ProfileVersion: profile.Version,
		State:          ConfigStateApplying,
		ManagedPaths:   append([]string(nil), profile.ManagedPaths...),
		GeneratedFiles: configFilePaths(profile),
		ManagedPlugins: append([]string(nil), profile.ManagedPlugins...),
	}
	if err := SaveConfigurationRecord(options.Paths, record); err != nil {
		result.State = ConfigStateFailed
		return result, err
	}
	progress(options, i18n.ServiceConfigProgress, map[string]any{"Action": "creating configuration files"})

	createdFiles := []string{}
	createdPaths := []string{}
	for _, path := range profile.ManagedPaths {
		if mkdir := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "mkdir", "-p", path); mkdir.Err != nil {
			return failConfigApply(ctx, options, record, createdFiles, createdPaths, result, mkdir.Err)
		}
		createdPaths = append(createdPaths, path)
	}
	for _, file := range profile.Files {
		encoded := base64.StdEncoding.EncodeToString([]byte(file.Content))
		mode := file.Mode
		if mode == 0 {
			mode = 0o600
		}
		write := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "sh", "-ec", configWriteScript, "--", file.Path, encoded, fmt.Sprintf("%o", mode))
		if write.Err != nil {
			return failConfigApply(ctx, options, record, createdFiles, createdPaths, result, write.Err)
		}
		createdFiles = append(createdFiles, file.Path)
	}
	for _, plugin := range profile.Plugins {
		progress(options, i18n.ServiceConfigPlugin, map[string]any{"Name": plugin.Name})
		clone := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "git", "clone", "--filter=blob:none", "--no-checkout", "--", plugin.Repository, plugin.Path)
		if clone.Err != nil {
			return failConfigApply(ctx, options, record, createdFiles, createdPaths, result, clone.Err)
		}
		createdPaths = append(createdPaths, plugin.Path)
		fetch := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "git", "-C", plugin.Path, "fetch", "--depth", "1", "origin", plugin.Commit)
		if fetch.Err != nil {
			return failConfigApply(ctx, options, record, createdFiles, createdPaths, result, fetch.Err)
		}
		checkout := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "git", "-C", plugin.Path, "checkout", "--detach", plugin.Commit)
		if checkout.Err != nil {
			return failConfigApply(ctx, options, record, createdFiles, createdPaths, result, checkout.Err)
		}
	}
	for _, validation := range profile.Validation {
		progress(options, i18n.ServiceConfigProgress, map[string]any{"Action": "validating configuration"})
		if command := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", validation.Name, validation.Args...); command.Err != nil {
			return failConfigApply(ctx, options, record, createdFiles, createdPaths, result, command.Err)
		}
	}

	hashes := map[string]string{}
	for _, path := range record.GeneratedFiles {
		hash, hashErr := currentFileHash(ctx, runner, options, "", path)
		if hashErr != nil {
			return failConfigApply(ctx, options, record, createdFiles, createdPaths, result, hashErr)
		}
		hashes[path] = hash
	}
	record.FileHashes = hashes
	record.State = ConfigStateApplied
	record.AppliedAt = options.Now().UTC()
	if err := SaveConfigurationRecord(options.Paths, record); err != nil {
		return result, err
	}
	result.State = ConfigStateApplied
	result.Success = true
	result.Changed = true
	result.Paths = append([]string(nil), record.ManagedPaths...)
	result.Message = options.Localizer.Text(i18n.OutputConfigApplied, nil)
	result.MessageID = string(i18n.OutputConfigApplied)
	return result, nil
}

func removeConfig(ctx context.Context, app string, options Options) (ConfigOperationResult, error) {
	profile, appProfile, result, err := resolveConfigProfile(app, options)
	if err != nil {
		return result, err
	}
	result.Action = "remove"
	result.Profile = profile.ID
	record, err := LoadConfigurationRecord(options.Paths, appProfile.Name)
	if err != nil {
		result.State = ConfigStateFailed
		return result, i18n.NewError(i18n.ServiceConfigError, "config_load", map[string]any{"Detail": fmt.Sprintf("%s", err)}, err)
	}
	if record.Profile != profile.ID {
		result.State = ConfigStateConflict
		return result, i18n.NewError(i18n.ServiceConfigError, "config_profile_owner", map[string]any{"Detail": profile.ID}, nil)
	}
	runner := runnerFor(options)
	progress(options, i18n.ServiceConfigProgress, map[string]any{"Action": "removing configuration"})
	record.State = ConfigStateRemoving
	if err := SaveConfigurationRecord(options.Paths, record); err != nil {
		return result, err
	}
	preserved := []string{}
	removed := []string{}
	for _, path := range record.GeneratedFiles {
		if err := validateConfigPath(path); err != nil {
			return failConfigRemove(options, record, preserved, result, err)
		}
		exists := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "test", "-e", path)
		if exists.Err != nil {
			continue
		}
		expected := record.FileHashes[path]
		current, hashErr := currentFileHash(ctx, runner, options, "", path)
		if hashErr != nil || expected == "" || current != expected {
			preserved = append(preserved, path)
			continue
		}
		if remove := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "rm", "--", path); remove.Err != nil {
			return failConfigRemove(options, record, preserved, result, remove.Err)
		}
		removed = append(removed, path)
	}
	for index := len(record.ManagedPaths) - 1; index >= 0; index-- {
		path := record.ManagedPaths[index]
		if err := validateConfigPath(path); err != nil {
			return failConfigRemove(options, record, preserved, result, err)
		}
		if remove := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "rmdir", "--", path); remove.Err != nil {
			preserved = append(preserved, path)
		}
	}
	for _, path := range record.ManagedPlugins {
		if err := validateConfigPath(path); err != nil {
			return failConfigRemove(options, record, preserved, result, err)
		}
		pluginStatus := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "git", "-C", path, "status", "--porcelain")
		if pluginStatus.Err != nil || len(strings.TrimSpace(string(pluginStatus.Stdout))) > 0 {
			preserved = append(preserved, path)
			continue
		}
		if remove := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "rm", "-rf", "--", path); remove.Err != nil {
			return failConfigRemove(options, record, preserved, result, remove.Err)
		}
		removed = append(removed, path)
	}
	record.ModifiedPaths = preserved
	record.RemovedAt = options.Now().UTC()
	record.State = ConfigStateRemoved
	if len(preserved) > 0 {
		record.State = ConfigStateModified
	}
	if err := SaveConfigurationRecord(options.Paths, record); err != nil {
		return result, err
	}
	result.State = record.State
	result.Success = true
	result.Changed = len(removed) > 0
	result.Paths = append([]string(nil), removed...)
	result.Conflicts = append([]string(nil), preserved...)
	result.Message = options.Localizer.Text(i18n.OutputConfigRemoved, nil)
	result.MessageID = string(i18n.OutputConfigRemoved)
	return result, nil
}

func resolveConfigProfile(app string, options Options) (ConfigProfile, AppProfile, ConfigOperationResult, error) {
	appProfile, ok := Resolve(app)
	result := ConfigOperationResult{SchemaVersion: 1, Command: "config", App: app, State: ConfigStateFailed}
	if !ok {
		return ConfigProfile{}, AppProfile{}, result, i18n.NewError(i18n.ServiceConfigError, "config_unsupported", map[string]any{"Detail": app}, nil)
	}
	result.App = appProfile.Name
	if appProfile.ConfigProfile == "" {
		result.State = ConfigStateUnavailable
		return ConfigProfile{}, appProfile, result, i18n.NewError(i18n.ServiceConfigError, "config_unavailable", map[string]any{"Detail": appProfile.Name}, nil)
	}
	profile, ok := options.ConfigProfiles[appProfile.ConfigProfile]
	if !ok {
		result.State = ConfigStateUnavailable
		return ConfigProfile{}, appProfile, result, i18n.NewError(i18n.ServiceConfigError, "config_profile_missing", map[string]any{"Detail": appProfile.ConfigProfile}, nil)
	}
	if profile.App != appProfile.Name {
		result.State = ConfigStateConflict
		return ConfigProfile{}, appProfile, result, i18n.NewError(i18n.ServiceConfigError, "config_profile_mismatch", map[string]any{"Detail": profile.ID}, nil)
	}
	result.StorageEstimate = profile.StorageEstimate
	return profile, appProfile, result, nil
}

func validateConfigProfile(profile ConfigProfile) error {
	if profile.ID == "" || profile.Version == "" || profile.App == "" {
		return i18n.NewError(i18n.ServiceConfigError, "config_profile_incomplete", nil, nil)
	}
	for _, path := range configPaths(profile) {
		if err := validateConfigPath(path); err != nil {
			return err
		}
	}
	for _, file := range profile.Files {
		if !configPathWithin(file.Path, profile.ManagedPaths) {
			return i18n.NewError(i18n.ServiceConfigError, "config_file_outside_paths", map[string]any{"Detail": file.Path}, nil)
		}
	}
	for _, plugin := range profile.Plugins {
		if plugin.Name == "" || plugin.Repository == "" || plugin.Commit == "" {
			return i18n.NewError(i18n.ServiceConfigError, "config_plugin_incomplete", nil, nil)
		}
		if !strings.HasPrefix(plugin.Repository, "https://") || len(plugin.Commit) != 40 {
			return i18n.NewError(i18n.ServiceConfigError, "config_plugin_unpinned", map[string]any{"Detail": plugin.Name}, nil)
		}
		if err := validateConfigPath(plugin.Path); err != nil {
			return err
		}
	}
	if len(profile.ManagedPlugins) != len(profile.Plugins) {
		return i18n.NewError(i18n.ServiceConfigError, "config_plugin_manifest", nil, nil)
	}
	return nil
}

func configPaths(profile ConfigProfile) []string {
	paths := append([]string(nil), profile.ManagedPaths...)
	for _, file := range profile.Files {
		paths = append(paths, file.Path)
	}
	for _, plugin := range profile.Plugins {
		paths = append(paths, plugin.Path)
	}
	return uniqueStrings(paths)
}

func configFilePaths(profile ConfigProfile) []string {
	paths := make([]string, 0, len(profile.Files))
	for _, file := range profile.Files {
		paths = append(paths, file.Path)
	}
	return paths
}

func configPathWithin(path string, parents []string) bool {
	for _, parent := range parents {
		relative, err := filepath.Rel(parent, path)
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && relative != "." {
			return true
		}
	}
	return false
}

func validateConfigPath(path string) error {
	clean := filepath.Clean(path)
	if path == "" || clean != path || !filepath.IsAbs(path) || (path != "/root" && !strings.HasPrefix(path, "/root/")) {
		return i18n.NewError(i18n.ServiceConfigError, "config_invalid_path", map[string]any{"Detail": path}, nil)
	}
	return nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func failConfigApply(ctx context.Context, options Options, record ConfigurationRecord, files, paths []string, result ConfigOperationResult, operationErr error) (ConfigOperationResult, error) {
	rollbackConfigAttempt(ctx, options, files, paths)
	record.State = ConfigStateFailed
	record.LastError = operationErr.Error()
	record.LastErrorCode = i18n.ErrorCode(operationErr)
	record.GeneratedFiles = files
	record.ManagedPaths = paths
	_ = SaveConfigurationRecord(options.Paths, record)
	result.State = ConfigStateFailed
	return result, operationErr
}

func rollbackConfigAttempt(ctx context.Context, options Options, files, paths []string) {
	runner := runnerFor(options)
	for _, path := range files {
		if validateConfigPath(path) == nil {
			_ = runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "rm", "--", path)
		}
	}
	for index := len(paths) - 1; index >= 0; index-- {
		if validateConfigPath(paths[index]) == nil {
			removed := runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "rmdir", "--", paths[index])
			if removed.Err != nil {
				_ = runUbuntuLogged(ctx, runner, commandTimeoutFor(options), "", "rm", "-rf", "--", paths[index])
			}
		}
	}
}

func failConfigRemove(options Options, record ConfigurationRecord, preserved []string, result ConfigOperationResult, operationErr error) (ConfigOperationResult, error) {
	record.State = ConfigStateFailed
	record.ModifiedPaths = preserved
	record.LastError = operationErr.Error()
	record.LastErrorCode = i18n.ErrorCode(operationErr)
	_ = SaveConfigurationRecord(options.Paths, record)
	result.State = ConfigStateFailed
	return result, operationErr
}

func runnerFor(options Options) CommandRunner {
	if options.Runner != nil {
		return options.Runner
	}
	return ExecRunner{}
}

func commandTimeoutFor(options Options) time.Duration {
	if options.CommandTimeout > 0 {
		return options.CommandTimeout
	}
	return defaultCommandTimeout
}

func configDefaults(options Options) Options {
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.CommandTimeout <= 0 {
		options.CommandTimeout = defaultCommandTimeout
	}
	if options.ConfigProfiles == nil {
		options.ConfigProfiles = DefaultConfigProfiles()
	}
	if options.Localizer.Locale == "" {
		options.Localizer = i18n.New(i18n.LocaleENUS)
	}
	return options
}
