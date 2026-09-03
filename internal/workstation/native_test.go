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

func TestSetupInstallsBasePackages(t *testing.T) {
	want := "install -y openssh net-tools wget iproute2 rsync dnsutils zip file jq ripgrep fd make git tree"
	if got := strings.Join(setupPackageArguments(), " "); got != want {
		t.Fatalf("setup package arguments = %q, want %q", got, want)
	}
}

func TestSetupPackagesPhaseRequiresCurrentVersion(t *testing.T) {
	p := paths.New(t.TempDir(), t.TempDir())
	if err := os.MkdirAll(p.StateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p.SetupPhase("packages-installed"), []byte(setupPackagesVersion+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !New(p).setupPackagesDone() {
		t.Fatal("current setup package version was not accepted")
	}
	if err := os.WriteFile(p.SetupPhase("packages-installed"), []byte("concluida\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if New(p).setupPackagesDone() {
		t.Fatal("legacy setup package marker was accepted")
	}
}
