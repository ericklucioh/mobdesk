package workstation

import (
	"os"
	"os/exec"
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
	config, err := os.ReadFile(p.ShellConfig())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(config), "Welcome to Mobdesk") {
		t.Fatal("managed shell configuration does not contain the SSH welcome banner")
	}
}

func TestShellBannerIsShownOnlyForInteractiveSSH(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is required to execute the generated shell configuration")
	}

	configPath := t.TempDir() + "/shell.bash"
	if err := os.WriteFile(configPath, []byte(renderShellConfig("test-version")), 0o600); err != nil {
		t.Fatal(err)
	}

	run := func(sshTTY, sshConnection string) string {
		command := exec.Command(bash, "--noprofile", "--norc", "-c", `source "$1"; exit`, "bash", configPath)
		command.Env = append(os.Environ(),
			"HOME="+t.TempDir(),
			"HOSTNAME=mobdesk-host",
			"USER=alice",
			"SSH_TTY="+sshTTY,
			"SSH_CONNECTION="+sshConnection,
		)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("execute generated shell configuration: %v\n%s", err, output)
		}
		return string(output)
	}

	sshOutput := run("/dev/pts/1", "192.0.2.10 45678 192.0.2.20 8022")
	for _, expected := range []string{"Welcome to Mobdesk", "test-version", "192.0.2.20", "alice", "8022", "Goodbye from Mobdesk"} {
		if !strings.Contains(sshOutput, expected) {
			t.Fatalf("SSH output does not contain %q: %q", expected, sshOutput)
		}
	}

	localOutput := run("", "")
	for _, unexpected := range []string{"Welcome to Mobdesk", "Goodbye from Mobdesk"} {
		if strings.Contains(localOutput, unexpected) {
			t.Fatalf("local output unexpectedly contains %q: %q", unexpected, localOutput)
		}
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
