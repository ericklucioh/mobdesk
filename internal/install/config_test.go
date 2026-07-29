package install

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type configRunner struct {
	existing map[string]bool
	hash     string
	commands []string
}

func (r *configRunner) Run(_ context.Context, name string, args ...string) CommandResult {
	r.commands = append(r.commands, name+" "+strings.Join(args, " "))
	tool := name
	for index, arg := range args {
		if strings.HasPrefix(arg, "PATH=") && index+1 < len(args) {
			tool = args[index+1]
			break
		}
	}
	if tool == "test" && len(args) > 1 && !r.existing[args[len(args)-1]] {
		return CommandResult{Err: errors.New("not found")}
	}
	if tool == "sha256sum" && len(args) > 1 {
		return CommandResult{Stdout: []byte(r.hash + "  " + args[len(args)-1] + "\n")}
	}
	return CommandResult{}
}

func testConfigOptions(t *testing.T, runner CommandRunner) Options {
	t.Helper()
	options := testOptions(t, runner)
	options.ConfigProfiles = map[string]ConfigProfile{
		"lazyvim": {
			ID:             "lazyvim",
			Version:        "1",
			App:            "neovim",
			Description:    "test profile",
			ManagedPaths:   []string{"/root/.config/nvim"},
			Files:          []ConfigFile{{Path: "/root/.config/nvim/init.lua", Content: "return {}"}},
			ConflictPolicy: "reject",
			Validation:     []ConfigCommand{{Name: "nvim", Args: []string{"--headless", "+qa"}}},
		},
	}
	writeInstallationRecord(t, options, InstallationRecord{Name: "neovim", Package: "neovim", State: "installed", Source: "mobdesk"})
	return options
}

func TestApplyConfigWritesPrivateRecordAndIsIdempotent(t *testing.T) {
	runner := &configRunner{hash: "hash"}
	options := testConfigOptions(t, runner)
	first, err := ApplyConfig(context.Background(), "nvim", options)
	if err != nil || !first.Success || first.State != ConfigStateApplied || !first.Changed {
		t.Fatalf("unexpected first apply: %+v, %v", first, err)
	}
	record, err := LoadConfigurationRecord(options.Paths, "neovim")
	if err != nil {
		t.Fatal(err)
	}
	if record.State != ConfigStateApplied || record.FileHashes["/root/.config/nvim/init.lua"] != "hash" {
		t.Fatalf("unexpected applied record: %+v", record)
	}
	second, err := ApplyConfig(context.Background(), "neovim", options)
	if err != nil || !second.Success || second.Changed || second.State != ConfigStateApplied {
		t.Fatalf("unexpected idempotent apply: %+v, %v", second, err)
	}
}

func TestApplyConfigRejectsExistingPathWithoutRecord(t *testing.T) {
	runner := &configRunner{hash: "hash", existing: map[string]bool{"/root/.config/nvim": true}}
	options := testConfigOptions(t, runner)
	result, err := ApplyConfig(context.Background(), "neovim", options)
	if err == nil || result.State != ConfigStateConflict || len(result.Conflicts) == 0 {
		t.Fatalf("unexpected conflict result: %+v, %v", result, err)
	}
	for _, command := range runner.commands {
		if strings.HasPrefix(command, "sh -ec ") {
			t.Fatalf("conflict executed file write: %s", command)
		}
	}
}

func TestRemoveConfigPreservesModifiedFile(t *testing.T) {
	path := "/root/.config/nvim/init.lua"
	runner := &configRunner{hash: "changed", existing: map[string]bool{"/root/.config/nvim": true, path: true}}
	options := testConfigOptions(t, runner)
	if err := SaveConfigurationRecord(options.Paths, ConfigurationRecord{
		App: "neovim", Profile: "lazyvim", ProfileVersion: "1", State: ConfigStateApplied,
		ManagedPaths: []string{"/root/.config/nvim"}, GeneratedFiles: []string{path}, FileHashes: map[string]string{path: "original"},
	}); err != nil {
		t.Fatal(err)
	}
	result, err := RemoveConfig(context.Background(), "neovim", options)
	if err != nil || result.State != ConfigStateModified || len(result.Conflicts) != 1 {
		t.Fatalf("unexpected modified removal: %+v, %v", result, err)
	}
}

func TestConfigRejectsUnsafePaths(t *testing.T) {
	for _, path := range []string{"", "/tmp/config", "/root/../etc/passwd", "relative"} {
		if err := validateConfigPath(path); err == nil {
			t.Fatalf("unsafe config path accepted: %q", path)
		}
	}
}
