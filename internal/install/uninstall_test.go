package install

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeInstallationRecord(t *testing.T, options Options, record InstallationRecord) {
	t.Helper()
	if err := os.MkdirAll(options.Paths.InstallationsDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := saveRecord(options.Paths.InstallationsDir(), record); err != nil {
		t.Fatal(err)
	}
}

func TestUninstallAptRemovesOnlyManagedPackage(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " apt-get -o DPkg::Lock::Timeout=300 remove -y neovim": {{}},
	}}
	options := testOptions(t, runner)
	writeInstallationRecord(t, options, InstallationRecord{
		Name: "neovim", Package: "neovim", Executable: "nvim", Strategy: "apt", Source: "mobdesk", State: "installed", LogPath: filepath.Join(options.Paths.InstallLogsDir(), "neovim.log"),
	})

	result, err := Uninstall(context.Background(), "nvim", options)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || result.State != "uninstalled" || result.Installed {
		t.Fatalf("unexpected uninstall result: %+v", result)
	}
	record, err := loadInstallationRecord(options.Paths, "neovim")
	if err != nil {
		t.Fatal(err)
	}
	if record.State != "uninstalled" || !reflect.DeepEqual(record.RemovedPackages, []string{"neovim"}) {
		t.Fatalf("unexpected uninstall record: %+v", record)
	}
	if strings.Contains(strings.Join(runner.commands, "\n"), "autoremove") {
		t.Fatalf("uninstall unexpectedly removed dependencies: %v", runner.commands)
	}
}

func TestUninstallRejectsDetectedInstallation(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{}}
	options := testOptions(t, runner)
	writeInstallationRecord(t, options, InstallationRecord{
		Name: "neovim", Package: "neovim", Strategy: "apt", Source: "detected", State: "installed",
	})
	_, err := Uninstall(context.Background(), "neovim", options)
	if err == nil || !strings.Contains(err.Error(), "apenas detectada") {
		t.Fatalf("unexpected detected uninstall result: %v", err)
	}
	if len(runner.commands) != 0 {
		t.Fatalf("detected uninstall executed commands: %v", runner.commands)
	}
}

func TestUninstallProtectsSharedPackage(t *testing.T) {
	runner := &fakeRunner{results: map[string][]CommandResult{}}
	options := testOptions(t, runner)
	writeInstallationRecord(t, options, InstallationRecord{Name: "c", Package: "clang", Strategy: "apt", Source: "mobdesk", State: "installed"})
	writeInstallationRecord(t, options, InstallationRecord{Name: "cpp", Package: "clang", Strategy: "apt", Source: "mobdesk", State: "installed"})
	_, err := Uninstall(context.Background(), "c", options)
	if err == nil || !strings.Contains(err.Error(), "pacote compartilhado") {
		t.Fatalf("unexpected shared package result: %v", err)
	}
	if len(runner.commands) != 0 {
		t.Fatalf("shared package uninstall executed commands: %v", runner.commands)
	}
}

func TestUninstallPreservesModifiedTrackedFile(t *testing.T) {
	path := "/root/.local/bin/yazi"
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " test -e " + path:      {{}},
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " sha256sum -- " + path: {{Stdout: []byte("actual  " + path + "\n")}},
	}}
	options := testOptions(t, runner)
	writeInstallationRecord(t, options, InstallationRecord{
		Name: "yazi", Strategy: "script", Source: "mobdesk", State: "installed", InstalledFiles: []string{path}, InstalledFileHashes: map[string]string{path: "expected"},
	})
	result, err := Uninstall(context.Background(), "yazi", options)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != "modified" {
		t.Fatalf("state = %q, want modified", result.State)
	}
	record, err := loadInstallationRecord(options.Paths, "yazi")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(record.PreservedFiles, []string{path}) || len(record.RemovedFiles) != 0 {
		t.Fatalf("modified file was not preserved: %+v", record)
	}
}

func TestUninstallRemovesUnmodifiedTrackedFile(t *testing.T) {
	path := "/usr/local/bin/lazygit"
	runner := &fakeRunner{results: map[string][]CommandResult{
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " test -e " + path:      {{}},
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " sha256sum -- " + path: {{Stdout: []byte("expected  " + path + "\n")}},
		"proot-distro login ubuntu -- env PATH=" + ubuntuPath + " rm -- " + path:        {{}},
	}}
	options := testOptions(t, runner)
	writeInstallationRecord(t, options, InstallationRecord{
		Name: "lazygit", Strategy: "script", Source: "mobdesk", State: "installed", InstalledFiles: []string{path}, InstalledFileHashes: map[string]string{path: "expected"},
	})
	result, err := Uninstall(context.Background(), "lazygit", options)
	if err != nil || result.State != "uninstalled" {
		t.Fatalf("unexpected file uninstall result: %+v, %v", result, err)
	}
}

func TestUninstallDoesNotAcceptInvalidTrackedPath(t *testing.T) {
	if err := validateManagedPath("/root/.local/bin/../secret"); err == nil {
		t.Fatal("invalid tracked path unexpectedly accepted")
	}
	if err := validateManagedPath("/tmp/file"); err == nil {
		t.Fatal("external tracked path unexpectedly accepted")
	}
}

func TestUninstallMissingRecord(t *testing.T) {
	options := testOptions(t, &fakeRunner{results: map[string][]CommandResult{}})
	_, err := Uninstall(context.Background(), "neovim", options)
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected missing record error: %v", err)
	}
}
