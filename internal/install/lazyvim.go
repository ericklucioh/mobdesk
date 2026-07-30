package install

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/ericklucioh/mobdesk/internal/i18n"
)

// The profile files are compiled into the binary so applying a profile does
// not depend on a mutable checkout or an unversioned remote template.
//
//go:embed profiles/neovim/*.lua profiles/neovim/*.json
var lazyVimFiles embed.FS

func profileFile(name string) string {
	content, err := lazyVimFiles.ReadFile("profiles/neovim/" + name)
	if err != nil {
		panic(fmt.Sprintf("LazyVim profile file missing: %s: %v", name, err))
	}
	return string(content)
}

// DefaultConfigProfiles returns the versioned configuration profiles shipped
// by Mobdesk. The returned map is safe for a caller to replace or extend.
func DefaultConfigProfiles(localizers ...i18n.Localizer) map[string]ConfigProfile {
	configRoot := "/root/.config/nvim"
	dataRoot := "/root/.local/share/nvim/lazy"
	plugins := []ConfigPlugin{
		{Name: "lazy.nvim", Repository: "https://github.com/folke/lazy.nvim.git", Commit: "306a05526ada86a7b30af95c5cc81ffba93fef97", Path: filepath.Join(dataRoot, "lazy.nvim")},
		{Name: "LazyVim", Repository: "https://github.com/LazyVim/LazyVim.git", Commit: "459a4c3b1059671e766a46c7cc223827dc67e3d0", Path: filepath.Join(dataRoot, "LazyVim")},
		{Name: "nvim-treesitter", Repository: "https://github.com/nvim-treesitter/nvim-treesitter.git", Commit: "61df84986b4b4ec469ee745a182e433d49f8c27e", Path: filepath.Join(dataRoot, "nvim-treesitter")},
	}
	managedPlugins := make([]string, 0, len(plugins))
	for _, plugin := range plugins {
		managedPlugins = append(managedPlugins, plugin.Path)
	}

	profiles := map[string]ConfigProfile{
		"lazyvim": {
			ID:            "lazyvim",
			Version:       "1.0.0",
			App:           "neovim",
			DescriptionID: i18n.ProfileLazyVimDescription,
			ManagedPaths: []string{
				configRoot,
				filepath.Join(configRoot, "lua"),
				filepath.Join(configRoot, "lua", "config"),
				filepath.Join(configRoot, "lua", "plugins"),
			},
			Files:          lazyVimConfigFiles(configRoot),
			Plugins:        plugins,
			ManagedPlugins: managedPlugins,
			Validation: []ConfigCommand{
				{Name: "nvim", Args: []string{"--headless", "+lua", "assert(pcall(require, 'lazy'))", "+qa"}},
			},
			ConflictPolicy:  "reject",
			StorageEstimate: plannedStorage(0, 0, 0, 20, 100, 300),
		},
	}
	// Keep the existing TUI presentation until its dedicated localization phase.
	localizer := i18n.New(i18n.LocalePTBR)
	if len(localizers) > 0 {
		localizer = localizers[0]
	}
	for id, profile := range profiles {
		profile.Description = localizer.Text(profile.DescriptionID, nil)
		profiles[id] = profile
	}
	return profiles
}

func lazyVimConfigFiles(configRoot string) []ConfigFile {
	files := []struct {
		name string
		path string
		mode uint32
	}{
		{name: "init.lua", path: filepath.Join(configRoot, "init.lua"), mode: 0o600},
		{name: "lazy.lua", path: filepath.Join(configRoot, "lua", "config", "lazy.lua"), mode: 0o600},
		{name: "mobdesk.lua", path: filepath.Join(configRoot, "lua", "plugins", "mobdesk.lua"), mode: 0o600},
		{name: "lazy-lock.json", path: filepath.Join(configRoot, "lazy-lock.json"), mode: 0o600},
	}
	result := make([]ConfigFile, 0, len(files))
	for _, file := range files {
		content := profileFile(file.name)
		if filepath.Ext(file.name) == ".json" {
			var document map[string]any
			if err := json.Unmarshal([]byte(content), &document); err != nil {
				panic(fmt.Sprintf("invalid LazyVim profile: %s: %v", file.name, err))
			}
		}
		result = append(result, ConfigFile{Path: file.path, Content: content, Mode: file.mode})
	}
	return result
}
