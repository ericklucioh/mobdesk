package install

import (
	"context"
	"strings"
	"testing"
)

func TestDefaultLazyVimProfileIsVersionedAndComplete(t *testing.T) {
	profile, ok := DefaultConfigProfiles()["lazyvim"]
	if !ok {
		t.Fatal("lazyvim profile is missing")
	}
	if err := validateConfigProfile(profile); err != nil {
		t.Fatal(err)
	}
	if profile.App != "neovim" || profile.Version == "" || len(profile.Files) != 4 || len(profile.Plugins) != 3 {
		t.Fatalf("unexpected LazyVim profile: %+v", profile)
	}
	for _, plugin := range profile.Plugins {
		if !strings.HasPrefix(plugin.Repository, "https://") || len(plugin.Commit) != 40 || plugin.Path == "" {
			t.Fatalf("plugin is not fixed: %+v", plugin)
		}
	}
	for _, file := range profile.Files {
		if file.Content == "" || !strings.HasPrefix(file.Path, "/root/.config/nvim/") {
			t.Fatalf("invalid embedded profile file: %+v", file)
		}
	}
	if profile.StorageEstimate == nil || profile.StorageEstimate.ConfigMinMB != 100 || profile.StorageEstimate.ConfigMaxMB != 300 {
		t.Fatalf("unexpected LazyVim estimate: %+v", profile.StorageEstimate)
	}
}

func TestApplyDefaultLazyVimClonesPinnedPlugins(t *testing.T) {
	runner := &configRunner{hash: "hash"}
	options := testOptions(t, runner)
	writeInstallationRecord(t, options, InstallationRecord{Name: "neovim", Package: "neovim", State: "installed", Source: "mobdesk"})

	result, err := ApplyConfig(context.Background(), "neovim", options)
	if err != nil || !result.Success || result.State != ConfigStateApplied {
		t.Fatalf("unexpected LazyVim apply: %+v, %v", result, err)
	}
	profile := DefaultConfigProfiles()["lazyvim"]
	for _, plugin := range profile.Plugins {
		if !containsCommand(runner.commands, "git clone --filter=blob:none --no-checkout -- "+plugin.Repository+" "+plugin.Path) || !containsCommand(runner.commands, "git -C "+plugin.Path+" checkout --detach "+plugin.Commit) {
			t.Fatalf("pinned plugin commands missing for %s: %v", plugin.Name, runner.commands)
		}
	}
	record, err := LoadConfigurationRecord(options.Paths, "neovim")
	if err != nil {
		t.Fatal(err)
	}
	if len(record.ManagedPlugins) != len(profile.Plugins) || len(record.GeneratedFiles) != 4 {
		t.Fatalf("plugin manifest was not persisted: %+v", record)
	}
}

func TestLazyVimProfileRejectsUnfixedPlugin(t *testing.T) {
	profile := DefaultConfigProfiles()["lazyvim"]
	profile.Plugins[0].Commit = "stable"
	if err := validateConfigProfile(profile); err == nil {
		t.Fatal("unfixed plugin revision was accepted")
	}
}

func TestRemoveDefaultLazyVimRemovesCleanManagedPlugins(t *testing.T) {
	runner := &configRunner{}
	options := testOptions(t, runner)
	profile := DefaultConfigProfiles()["lazyvim"]
	managedPlugins := make([]string, 0, len(profile.Plugins))
	for _, plugin := range profile.Plugins {
		managedPlugins = append(managedPlugins, plugin.Path)
	}
	writeInstallationRecord(t, options, InstallationRecord{Name: "neovim", Package: "neovim", State: "installed", Source: "mobdesk"})
	if err := SaveConfigurationRecord(options.Paths, ConfigurationRecord{
		App: "neovim", Profile: "lazyvim", ProfileVersion: profile.Version,
		State: ConfigStateApplied, ManagedPlugins: managedPlugins,
	}); err != nil {
		t.Fatal(err)
	}

	result, err := RemoveConfig(context.Background(), "nvim", options)
	if err != nil || result.State != ConfigStateRemoved || len(result.Conflicts) != 0 {
		t.Fatalf("unexpected LazyVim removal: %+v, %v", result, err)
	}
	for _, path := range managedPlugins {
		if !containsCommand(runner.commands, "rm -rf -- "+path) {
			t.Fatalf("managed plugin was not removed: %s; commands=%v", path, runner.commands)
		}
	}
}

func TestRemoveDefaultLazyVimPreservesModifiedPlugin(t *testing.T) {
	runner := &configRunner{pluginStatus: " M lua/init.lua\n"}
	options := testOptions(t, runner)
	profile := DefaultConfigProfiles()["lazyvim"]
	managedPlugins := make([]string, 0, len(profile.Plugins))
	for _, plugin := range profile.Plugins {
		managedPlugins = append(managedPlugins, plugin.Path)
	}
	writeInstallationRecord(t, options, InstallationRecord{Name: "neovim", Package: "neovim", State: "installed", Source: "mobdesk"})
	if err := SaveConfigurationRecord(options.Paths, ConfigurationRecord{
		App: "neovim", Profile: "lazyvim", ProfileVersion: profile.Version,
		State: ConfigStateApplied, ManagedPlugins: managedPlugins,
	}); err != nil {
		t.Fatal(err)
	}

	result, err := RemoveConfig(context.Background(), "neovim", options)
	if err != nil || result.State != ConfigStateModified || len(result.Conflicts) != len(managedPlugins) {
		t.Fatalf("unexpected modified plugin removal: %+v, %v", result, err)
	}
	for _, path := range managedPlugins {
		if containsCommand(runner.commands, "rm -rf -- "+path) {
			t.Fatalf("modified plugin was removed: %s; commands=%v", path, runner.commands)
		}
	}
}

func containsCommand(commands []string, wanted string) bool {
	for _, command := range commands {
		if strings.Contains(command, wanted) {
			return true
		}
	}
	return false
}
