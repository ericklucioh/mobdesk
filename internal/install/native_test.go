package install

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ericklucioh/mobdesk/internal/paths"
)

type nativeRunner struct {
	commands []string
	versions int
}

func (r *nativeRunner) Run(_ context.Context, name string, args ...string) CommandResult {
	r.commands = append(r.commands, name+" "+strings.Join(args, " "))
	if name == "git" {
		r.versions++
		if r.versions == 1 {
			return CommandResult{Err: errors.New("git missing")}
		}
		return CommandResult{Stdout: []byte("git version 1.0\n")}
	}
	return CommandResult{}
}

func TestInstallUsesNativePkg(t *testing.T) {
	runner := &nativeRunner{}
	p := paths.New(t.TempDir(), "")
	result, err := Install(context.Background(), "git", Options{
		Paths:       p,
		Runner:      runner,
		Now:         time.Now,
		StorageFree: func(string) (int64, error) { return StorageWarningBytes + 1, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Installed || !result.Changed {
		t.Fatalf("unexpected installation result: %+v", result)
	}
	if !containsCommand(runner.commands, "pkg install -y git") {
		t.Fatalf("native pkg command was not used: %v", runner.commands)
	}
}

func TestCatalogUsesOnlyNativePkgProfiles(t *testing.T) {
	want := map[string]bool{
		"git": true, "neovim": true, "tmux": true, "go": true, "python": true,
		"node": true, "c": true, "cpp": true, "lua": true, "gh": true,
		"tree": true, "htop": true, "ncdu": true, "micro": true,
	}
	for _, profile := range Tools() {
		if !want[profile.Name] {
			t.Fatalf("unexpected catalog profile %q", profile.Name)
		}
		delete(want, profile.Name)
		if profile.InstallKind != "pkg" || profile.Package == "" || profile.Executable == "" {
			t.Fatalf("profile %q is not a complete native pkg profile: %+v", profile.Name, profile)
		}
	}
	if len(want) > 0 {
		t.Fatalf("missing catalog profiles: %v", want)
	}
}

func TestSharedNativePackageBlocksRemoval(t *testing.T) {
	p := paths.New(t.TempDir(), "")
	if err := os.MkdirAll(p.InstallationsDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := saveRecord(p.InstallationsDir(), InstallationRecord{Name: "c", Packages: []string{"clang"}, State: "installed"}); err != nil {
		t.Fatal(err)
	}
	if err := saveRecord(p.InstallationsDir(), InstallationRecord{Name: "cpp", Packages: []string{"clang"}, State: "installed"}); err != nil {
		t.Fatal(err)
	}
	if !packageSharedByAnotherInstallation(p, InstallationRecord{Name: "c", Packages: []string{"clang"}}) {
		t.Fatal("shared clang package was not detected")
	}
}

func containsCommand(commands []string, want string) bool {
	for _, command := range commands {
		if command == want {
			return true
		}
	}
	return false
}
