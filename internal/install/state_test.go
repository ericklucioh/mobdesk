package install

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/ericklucioh/mobdesk/internal/paths"
)

func TestConfigurationRecordUsesPrivateAtomicState(t *testing.T) {
	p := paths.New(t.TempDir(), "/termux")
	record := ConfigurationRecord{
		App:            "neovim",
		Profile:        "lazyvim",
		ProfileVersion: "1",
		State:          ConfigStateNotApplied,
		ManagedPaths:   []string{"/root/.config/nvim"},
		AppliedAt:      time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC),
	}
	if err := SaveConfigurationRecord(p, record); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadConfigurationRecord(p, "neovim")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded, record) {
		t.Fatalf("loaded record = %+v, want %+v", loaded, record)
	}
	directoryInfo, err := os.Stat(p.ConfigurationsDir())
	if err != nil {
		t.Fatal(err)
	}
	if directoryInfo.Mode().Perm() != 0o700 {
		t.Fatalf("configuration directory mode = %o, want 700", directoryInfo.Mode().Perm())
	}
	fileInfo, err := os.Stat(p.ConfigurationState("neovim"))
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Mode().Perm() != 0o600 {
		t.Fatalf("configuration state mode = %o, want 600", fileInfo.Mode().Perm())
	}
	if _, err := os.Stat(filepath.Join(p.ConfigurationsDir(), ".state-neovim.tmp")); !os.IsNotExist(err) {
		t.Fatalf("temporary state file was not cleaned up: %v", err)
	}
}

func TestConfigurationRecordRejectsPathEscapes(t *testing.T) {
	p := paths.New(t.TempDir(), "/termux")
	for _, app := range []string{"", ".", "..", "../escape", "nested/app", "/tmp/app"} {
		t.Run(app, func(t *testing.T) {
			err := SaveConfigurationRecord(p, ConfigurationRecord{App: app})
			if err == nil {
				t.Fatalf("SaveConfigurationRecord(%q) unexpectedly succeeded", app)
			}
		})
	}
}

func TestDeclaredInstalledFilesRespectProfileLocation(t *testing.T) {
	if got := declaredInstalledFiles(AppProfile{Name: "neovim", Executable: "nvim", InstallKind: "apt"}); got != nil {
		t.Fatalf("apt files = %v, want nil", got)
	}
	if got := declaredInstalledFiles(AppProfile{Name: "yazi", Executable: "yazi", InstallKind: "script", UserBin: true}); !reflect.DeepEqual(got, []string{"/root/.local/bin/yazi", "/root/.local/bin/ya"}) {
		t.Fatalf("yazi files = %v", got)
	}
}
