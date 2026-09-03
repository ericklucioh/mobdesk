package install

import (
	"strings"
	"time"

	"github.com/ericklucioh/mobdesk/internal/i18n"
)

var catalogEstimateMeasuredAt = time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)

// catalog contains native Termux packages and individually audited user CLIs
// installed into private Mobdesk-owned directories.
var catalog = []AppProfile{
	{Name: "neovim", Aliases: []string{"nvim"}, DescriptionID: i18n.AppNeovimDescription, Usage: "nvim [file or directory]", Package: "neovim", Executable: "nvim", VersionArg: []string{"--version"}, Kind: "editor", InstallKind: "pkg", StorageEstimate: plannedStorage(15, 30, 0, 20)},
	{Name: "tmux", DescriptionID: i18n.AppTmuxDescription, Usage: "tmux [command]", Package: "tmux", Executable: "tmux", VersionArg: []string{"-V"}, Kind: "terminal", InstallKind: "pkg", StorageEstimate: plannedStorage(2, 5, 0, 2)},
	{Name: "go", Aliases: []string{"golang"}, DescriptionID: i18n.AppGoDescription, Usage: "go [command]", Package: "golang", Executable: "go", VersionArg: []string{"version"}, Kind: "language", InstallKind: "pkg", StorageEstimate: plannedStorage(180, 300, 0, 50)},
	{Name: "java", Aliases: []string{"jdk", "openjdk"}, DescriptionID: i18n.AppJavaDescription, Usage: "java [options] <class>", Package: "openjdk-21", Executable: "java", VersionArg: []string{"--version"}, RequiredExecutables: []ExecutableSpec{{Name: "java", VersionArg: []string{"--version"}}, {Name: "javac", VersionArg: []string{"--version"}}, {Name: "jar", VersionArg: []string{"--version"}}}, Kind: "language", InstallKind: "pkg", StorageEstimate: plannedStorage(257, 360, 50, 150)},
	{Name: "maven", Aliases: []string{"mvn"}, DescriptionID: i18n.AppMavenDescription, Usage: "mvn [goal]", Package: "maven", Executable: "mvn", VersionArg: []string{"--version"}, Kind: "development", InstallKind: "pkg", Requires: []string{"java"}, StorageEstimate: plannedStorage(9, 15, 0, 0)},
	{Name: "kotlin", Aliases: []string{"kotlin-jvm", "kotlinc"}, DescriptionID: i18n.AppKotlinDescription, Usage: "kotlinc [options] <source files>", Package: "kotlin", Executable: "kotlinc", VersionArg: []string{"-version"}, RequiredExecutables: []ExecutableSpec{{Name: "kotlinc", VersionArg: []string{"-version"}}, {Name: "kotlin", VersionArg: []string{"-version"}}}, Kind: "language", InstallKind: "pkg", Requires: []string{"java"}, StorageEstimate: plannedStorage(85, 110, 0, 0)},
	{Name: "gradle", Aliases: []string{"gradle-build"}, DescriptionID: i18n.AppGradleDescription, Usage: "gradle [options] [tasks...]", Package: "gradle", Executable: "gradle", VersionArg: []string{"--version"}, Kind: "development", InstallKind: "pkg", Requires: []string{"java"}, StorageEstimate: plannedStorage(140, 190, 0, 0)},
	{Name: "python", Aliases: []string{"python3"}, DescriptionID: i18n.AppPythonDescription, Usage: "python [script]", Package: "python", Executable: "python", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "pkg", StorageEstimate: plannedStorage(35, 60, 0, 20)},
	{Name: "node", Aliases: []string{"nodejs"}, DescriptionID: i18n.AppNodeDescription, Usage: "node [script]", Package: "nodejs", Executable: "node", VersionArg: []string{"--version"}, RequiredExecutables: []ExecutableSpec{{Name: "node", VersionArg: []string{"--version"}}, {Name: "npm", VersionArg: []string{"--version"}}}, Kind: "language", InstallKind: "pkg", StorageEstimate: plannedStorage(70, 130, 20, 60)},
	{Name: "c", Aliases: []string{"c-lang"}, DescriptionID: i18n.AppCDescription, Usage: "clang [options] <files...>", Package: "clang", Executable: "clang", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "pkg", StorageEstimate: plannedStorage(250, 450, 20, 80)},
	{Name: "cpp", Aliases: []string{"c++", "cplusplus"}, DescriptionID: i18n.AppCPPDescription, Usage: "clang++ [options] <files...>", Package: "clang", Executable: "clang++", VersionArg: []string{"--version"}, Kind: "language", InstallKind: "pkg", StorageEstimate: plannedStorage(250, 450, 20, 80)},
	{Name: "lua", Aliases: []string{"lua5.4"}, DescriptionID: i18n.AppLuaDescription, Usage: "lua [script]", Package: "lua54", Executable: "lua", VersionArg: []string{"-v"}, Kind: "language", InstallKind: "pkg", StorageEstimate: plannedStorage(1, 3, 0, 1)},
	{Name: "gh", Aliases: []string{"github-cli"}, DescriptionID: i18n.AppGHDescription, Usage: "gh <command> [options]", Package: "gh", Executable: "gh", VersionArg: []string{"--version"}, Kind: "development", InstallKind: "pkg", StorageEstimate: plannedStorage(10, 20, 0, 5)},
	{Name: "zellij", DescriptionID: i18n.AppZellijDescription, Usage: "zellij [options]", Package: "zellij", Executable: "zellij", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "pkg", StorageEstimate: plannedStorage(20, 30, 0, 5)},
	{Name: "lazygit", DescriptionID: i18n.AppLazygitDescription, Usage: "lazygit [directory]", Package: "lazygit", Executable: "lazygit", VersionArg: []string{"--version"}, Kind: "development", InstallKind: "pkg", StorageEstimate: plannedStorage(15, 25, 0, 5)},
	{Name: "sqlite", Aliases: []string{"sqlite3"}, DescriptionID: i18n.AppSQLiteDescription, Usage: "sqlite3 [database]", Package: "sqlite", Executable: "sqlite3", VersionArg: []string{"--version"}, Kind: "database", InstallKind: "pkg", StorageEstimate: plannedStorage(2, 5, 0, 2)},
	{Name: "mariadb", Aliases: []string{"mysql"}, DescriptionID: i18n.AppMariaDBDescription, Usage: "mariadb [options] [database]", Package: "mariadb", Executable: "mariadb", VersionArg: []string{"--version"}, Kind: "database", InstallKind: "pkg", StorageEstimate: plannedStorage(160, 280, 0, 20)},
	{Name: "postgresql", Aliases: []string{"postgres", "psql"}, DescriptionID: i18n.AppPostgreSQLDescription, Usage: "psql [options] [database]", Package: "postgresql", Executable: "psql", VersionArg: []string{"--version"}, Kind: "database", InstallKind: "pkg", StorageEstimate: plannedStorage(45, 100, 0, 20)},
	{Name: "htop", DescriptionID: i18n.AppHtopDescription, Usage: "htop", Package: "htop", Executable: "htop", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "pkg", StorageEstimate: plannedStorage(1, 3, 0, 1)},
	{Name: "ncdu", DescriptionID: i18n.AppNcduDescription, Usage: "ncdu [directory]", Package: "ncdu", Executable: "ncdu", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "pkg", StorageEstimate: plannedStorage(1, 2, 0, 1)},
	{Name: "inxi", DescriptionID: i18n.AppInxiDescription, Usage: "inxi [options]", Package: "inxi", Executable: "inxi", VersionArg: []string{"--version"}, Kind: "monitoring", InstallKind: "pkg", StorageEstimate: plannedStorage(5, 15, 0, 5)},
	{Name: "yazi", Aliases: []string{"yazi-fm"}, DescriptionID: i18n.AppYaziDescription, Usage: "yazi [directory]", Package: "yazi", Executable: "yazi", VersionArg: []string{"--version"}, Kind: "file", InstallKind: "pkg", StorageEstimate: plannedStorage(25, 40, 0, 0)},
	{Name: "micro", DescriptionID: i18n.AppMicroDescription, Usage: "micro [files...]", Package: "micro", Executable: "micro", VersionArg: []string{"--version"}, Kind: "terminal", InstallKind: "pkg", StorageEstimate: plannedStorage(4, 8, 0, 2)},
	{Name: "rclone", DescriptionID: i18n.AppRcloneDescription, Usage: "rclone <command> [options]", Package: "rclone", Executable: "rclone", VersionArg: []string{"version"}, Kind: "file", InstallKind: "pkg", StorageEstimate: plannedStorage(25, 45, 0, 10)},
	{
		Name:            "tuifi",
		Aliases:         []string{"tuifimanager"},
		DescriptionID:   i18n.AppTuifiDescription,
		Usage:           "tuifi [directory]",
		Package:         "TUIFIManager==5.2.6",
		Executable:      "tuifi",
		VersionArg:      []string{"--version"},
		Kind:            "file",
		InstallKind:     "pipx",
		Requires:        []string{"python"},
		UserBin:         true,
		StorageEstimate: plannedStorage(20, 40, 90, 180),
	},
	{
		Name:            "bitwarden",
		Aliases:         []string{"bw"},
		DescriptionID:   i18n.AppBitwardenDescription,
		Usage:           "bw <command> [options]",
		Package:         "@bitwarden/cli@2025.12.0",
		Executable:      "bw",
		VersionArg:      []string{"--version"},
		Kind:            "security",
		InstallKind:     "npm",
		Requires:        []string{"node"},
		UserBin:         true,
		StorageEstimate: plannedStorage(15, 30, 40, 100),
	},
	{
		Name:            "pi",
		Aliases:         []string{"pi-coding-agent"},
		DescriptionID:   i18n.AppCodexDescription,
		Usage:           "pi [options] [@files...] [messages...]",
		Package:         "@earendil-works/pi-coding-agent@0.84.4",
		Executable:      "pi",
		VersionArg:      []string{"--version"},
		Kind:            "development",
		InstallKind:     "npm",
		Requires:        []string{"node"},
		UserBin:         true,
		StorageEstimate: plannedStorage(80, 180, 100, 250),
	},
	{
		Name:            "resterm",
		DescriptionID:   i18n.AppRestermDescription,
		Usage:           "resterm [file or directory]",
		Package:         "github.com/unkn0wn-root/resterm/cmd/resterm@v1.2.1",
		Executable:      "resterm",
		VersionArg:      []string{"--version"},
		Kind:            "development",
		InstallKind:     "go",
		Requires:        []string{"go"},
		UserBin:         true,
		StorageEstimate: plannedStorage(85, 100, 100, 300),
	},
	{
		Name:            "ttt",
		DescriptionID:   i18n.AppTTTDescription,
		Usage:           "ttt [file or directory]",
		Package:         "github.com/eugenioenko/ttt/cmd/ttt@v1.1.0",
		Executable:      "ttt",
		VersionArg:      []string{"--version"},
		Kind:            "editor",
		InstallKind:     "go",
		Requires:        []string{"go"},
		UserBin:         true,
		StorageEstimate: plannedStorage(25, 45, 100, 350),
	},
}

func plannedStorage(appMin, appMax, dependenciesMin, dependenciesMax int64) *StorageEstimate {
	return &StorageEstimate{
		AppMinMB:          appMin,
		AppMaxMB:          appMax,
		DependenciesMinMB: dependenciesMin,
		DependenciesMaxMB: dependenciesMax,
		Source:            "planning",
		Version:           "termux-first",
		Architecture:      "arm64",
		MeasuredAt:        catalogEstimateMeasuredAt,
	}
}

// Catalog returns a localized copy of the curated application profiles.
func Catalog(localizers ...i18n.Localizer) []AppProfile {
	profiles := append([]AppProfile(nil), catalog...)
	localizer := i18n.New(i18n.LocalePTBR)
	if len(localizers) > 0 {
		localizer = localizers[0]
	}
	for index := range profiles {
		profiles[index].Description = localizer.Text(profiles[index].DescriptionID, nil)
	}
	return profiles
}

// Resolve returns the profile identified by its name or alias.
func Resolve(name string) (AppProfile, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, profile := range catalog {
		if name == profile.Name {
			return profile, true
		}
		for _, alias := range profile.Aliases {
			if name == alias {
				return profile, true
			}
		}
	}
	return AppProfile{}, false
}
