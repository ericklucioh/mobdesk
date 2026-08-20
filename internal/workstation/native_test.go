package workstation

import (
	"os"
	"strings"
	"testing"

	"github.com/ericklucioh/mobdesk/internal/paths"
)

func TestSSHConfigUsesNativeTermuxShell(t *testing.T) {
	config := renderSSHConfig(paths.New("/home/user", "/termux"))
	if strings.Contains(config, "ForceCommand") {
		t.Fatalf("native SSH configuration must not force a secondary shell: %s", config)
	}
}

func TestConfigureShellAddsOneManagedSourceBlock(t *testing.T) {
	p := paths.New(t.TempDir(), t.TempDir())
	service := New(p)
	if err := os.MkdirAll(p.ConfigDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := service.configureShell(); err != nil {
		t.Fatal(err)
	}
	if err := service.configureShell(); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(p.Home + "/.bashrc")
	if err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(string(contents), "# >>> mobdesk >>>"); count != 1 {
		t.Fatalf("managed source block count = %d", count)
	}
}
