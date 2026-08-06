package workstation

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ericklucioh/mobdesk/internal/paths"
)

func TestSetupOrchestratesAllPhasesWithExplicitPaths(t *testing.T) {
	p := paths.New(t.TempDir(), t.TempDir())
	service := New(p)
	var commands []string
	service.Deps.Run = func(_ context.Context, name string, args ...string) error {
		commands = append(commands, name+" "+strings.Join(args, " "))
		if strings.Join(args, " ") == "login ubuntu -- true" {
			return errors.New("not installed")
		}
		return nil
	}
	configured := false
	service.Deps.EnsureSSHConfigured = func(got paths.Paths) error { configured = got == p; return nil }
	service.Deps.AndroidTimezone = func(context.Context) string { return "America/Sao_Paulo" }
	service.Deps.Executable = func() (string, error) { return "/bin/mobdesk", nil }
	service.Deps.Abs = func(path string) (string, error) { return path, nil }
	service.Deps.EvalSymlinks = func(path string) (string, error) { return path, nil }

	result, err := service.Setup(context.Background(), SetupOptions{UpgradeSystem: true, AllowPasswordPrompt: true})
	if err != nil {
		t.Fatal(err)
	}
	if !configured || len(result.Phases) != 10 {
		t.Fatalf("setup incompleto: configured=%t phases=%v", configured, result.Phases)
	}
	for _, phase := range result.Phases {
		if _, err := os.Stat(p.SetupPhase(phase)); err != nil {
			t.Fatalf("marker %s missing: %v", phase, err)
		}
	}
	if _, err := os.Stat(p.SetupDone()); err != nil {
		t.Fatalf("setup.done missing: %v", err)
	}
	wantPrefix := []string{"pkg update", "pkg upgrade -y -o Dpkg::Options::=--force-confold", "pkg install -y -o Dpkg::Options::=--force-confold proot-distro openssh net-tools", "proot-distro login ubuntu -- true", "proot-distro install ubuntu", "proot-distro login ubuntu -- sh -ec " + ubuntuTimezoneScript + " -- America/Sao_Paulo", "proot-distro login ubuntu -- mkdir -p /root/workspace /root/.config/mobdesk /root/.local/share/mobdesk", "passwd "}
	if len(commands) != len(wantPrefix)+4 || strings.Join(commands[:len(wantPrefix)], "\n") != strings.Join(wantPrefix, "\n") {
		t.Fatalf("ordem de comandos inesperada:\n%v", commands)
	}
	if commands[8] != "proot-distro login ubuntu -- dpkg --configure -a" {
		t.Fatalf("unexpected dpkg repair: %q", commands[8])
	}
	if commands[9] != "proot-distro login ubuntu -- apt-get -y update" {
		t.Fatalf("unexpected Ubuntu package update: %q", commands[9])
	}
	if commands[10] != "proot-distro login ubuntu -- apt-get -o DPkg::Lock::Timeout=300 install -y bash-completion" {
		t.Fatalf("unexpected completion installation: %q", commands[10])
	}
	if !strings.Contains(commands[11], "bash_completion") || !strings.Contains(commands[11], "command -v javac") || !strings.Contains(commands[11], "JAVA_HOME") || !strings.Contains(commands[11], "PATH=\"$JAVA_HOME/bin:$PATH\"") || !strings.Contains(commands[11], "CGO_ENABLED=0") || !strings.Contains(commands[11], "PS1=") {
		t.Fatalf("shell configuration missing: %q", commands[11])
	}
}

func TestRenderUbuntuShellConfigPreservesUserBashrcAndUsesDynamicJDKPath(t *testing.T) {
	config := renderUbuntuShellConfig(paths.New("/termux/home", "/termux/prefix"))
	for _, fragment := range []string{
		`[ "$HOME/.bashrc" != "/root/.config/mobdesk/bashrc" ]`,
		`readlink -f "$mobdesk_javac"`,
		`JAVA_HOME=${mobdesk_javac%%/bin/javac}`,
		`export PATH="$JAVA_HOME/bin:$PATH"`,
	} {
		if !strings.Contains(config, fragment) {
			t.Fatalf("generated shell config missing %q:\n%s", fragment, config)
		}
	}
	if strings.Contains(config, "/termux/prefix") || strings.Contains(config, "PREFIX=") {
		t.Fatalf("generated shell config leaked Termux environment: %s", config)
	}
}

func TestSetupReconcilesShellConfigForExistingSetup(t *testing.T) {
	p := paths.New(t.TempDir(), t.TempDir())
	service := New(p)
	reconciled := false
	service.Deps.Run = func(_ context.Context, name string, args ...string) error {
		if name == "proot-distro" && strings.Join(args, " ") == "login ubuntu -- sh -ec "+renderUbuntuShellConfig(p) {
			reconciled = true
			return nil
		}
		return nil
	}
	service.Deps.EnsureSSHConfigured = func(paths.Paths) error { return nil }
	service.Deps.AndroidTimezone = func(context.Context) string { return "America/Sao_Paulo" }
	service.Deps.Executable = func() (string, error) { return "/bin/mobdesk", nil }
	service.Deps.Abs = func(path string) (string, error) { return path, nil }
	service.Deps.EvalSymlinks = func(path string) (string, error) { return path, nil }

	if err := os.MkdirAll(p.StateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"directories", "packages-updated", "packages-installed", "ubuntu-installed", "workspace-created", "password-configured", "ssh-configured", "shell-configured", "launcher-installed"} {
		if err := os.WriteFile(p.SetupPhase(phase), []byte("done"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := service.Setup(context.Background(), SetupOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p.SetupDone()); err != nil {
		t.Fatalf("setup.done missing: %v", err)
	}
	if !reconciled {
		t.Fatal("existing shell configuration was not reconciled")
	}
}

func TestSetupUsesNonInteractiveAPTForHeadlessMode(t *testing.T) {
	p := paths.New(t.TempDir(), t.TempDir())
	service := New(p)
	var commands []string
	service.Deps.Run = func(_ context.Context, name string, args ...string) error {
		commands = append(commands, name+" "+strings.Join(args, " "))
		return nil
	}
	service.Deps.EnsureSSHConfigured = func(paths.Paths) error { return nil }
	service.Deps.AndroidTimezone = func(context.Context) string { return "America/Sao_Paulo" }
	service.Deps.Executable = func() (string, error) { return "/bin/mobdesk", nil }
	service.Deps.Abs = func(path string) (string, error) { return path, nil }
	service.Deps.EvalSymlinks = func(path string) (string, error) { return path, nil }

	if err := os.MkdirAll(p.StateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"directories", "packages-updated", "packages-installed", "ubuntu-installed", "workspace-created", "password-configured", "ssh-configured"} {
		if err := os.WriteFile(p.SetupPhase(phase), []byte("done"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := service.Setup(context.Background(), SetupOptions{}); err != nil {
		t.Fatal(err)
	}
	if len(commands) < 4 || commands[0] != "proot-distro login ubuntu -- sh -ec "+ubuntuTimezoneScript+" -- America/Sao_Paulo" || commands[1] != "proot-distro login ubuntu -- env DEBIAN_FRONTEND=noninteractive TZ=America/Sao_Paulo dpkg --configure -a" || commands[2] != "proot-distro login ubuntu -- apt-get -y update" || commands[3] != "proot-distro login ubuntu -- env DEBIAN_FRONTEND=noninteractive TZ=America/Sao_Paulo apt-get -o DPkg::Lock::Timeout=300 install -y bash-completion" {
		t.Fatalf("headless setup did not configure APT non-interactively: %v", commands)
	}
}

func TestSetupCreatesPrivateStateDirectories(t *testing.T) {
	p := paths.New(t.TempDir(), t.TempDir())
	service := New(p)
	service.Deps.Run = func(_ context.Context, _ string, _ ...string) error { return nil }
	service.Deps.EnsureSSHConfigured = func(paths.Paths) error { return nil }
	service.Deps.Executable = func() (string, error) { return "/bin/mobdesk", nil }
	service.Deps.Abs = func(path string) (string, error) { return path, nil }
	service.Deps.EvalSymlinks = func(path string) (string, error) { return path, nil }

	if _, err := service.Setup(context.Background(), SetupOptions{AllowPasswordPrompt: true}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p.StateDir())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("state mode = %o, want 700", info.Mode().Perm())
	}
	marker, err := os.Stat(p.SetupPhase("directories"))
	if err != nil {
		t.Fatal(err)
	}
	if marker.Mode().Perm() != 0o600 {
		t.Fatalf("marker mode = %o, want 600", marker.Mode().Perm())
	}
}

func TestSetupRefusesSymlinkedPhaseMarker(t *testing.T) {
	p := paths.New(t.TempDir(), t.TempDir())
	service := New(p)
	if err := os.MkdirAll(p.StateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	target := p.DataDir() + "/outside"
	if err := os.WriteFile(target, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, p.SetupPhase("directories")); err != nil {
		t.Fatal(err)
	}

	err := service.markSetupPhase("directories")
	if err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("unexpected error: %v", err)
	}
	contents, readErr := os.ReadFile(target)
	if readErr != nil || string(contents) != "keep" {
		t.Fatalf("target = %q, err = %v", contents, readErr)
	}
}

func TestValidTimezoneRejectsUnsafePaths(t *testing.T) {
	for _, zone := range []string{"America/Sao_Paulo", "Etc/UTC", "UTC"} {
		if !validTimezone(zone) {
			t.Fatalf("validTimezone(%q) = false", zone)
		}
	}
	for _, zone := range []string{"", "/etc/passwd", "../UTC", "America/../UTC", "America/Sao Paulo"} {
		if validTimezone(zone) {
			t.Fatalf("validTimezone(%q) = true", zone)
		}
	}
}

func TestConfigureUbuntuTimezoneRejectsInvalidValue(t *testing.T) {
	service := New(paths.New(t.TempDir(), t.TempDir()))
	service.Deps.AndroidTimezone = func(context.Context) string { return "../../etc/passwd" }
	service.Deps.Run = func(context.Context, string, ...string) error {
		t.Fatal("invalid timezone reached Ubuntu")
		return nil
	}
	if err := service.configureUbuntuTimezone(context.Background()); err == nil {
		t.Fatal("invalid timezone was accepted")
	}
}
