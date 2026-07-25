package workstation

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ericklucioh/mobdesk/internal/paths"
)

func TestStartOrchestratesWorkstationWithoutSystemCommands(t *testing.T) {
	p := paths.New(t.TempDir(), t.TempDir())
	for _, marker := range []string{p.SetupDone(), p.PasswordDone()} {
		if err := os.MkdirAll(filepath.Dir(marker), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(marker, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	service := New(p)
	var ran, configured, locked, started bool
	service.Deps.Run = func(_ context.Context, name string, args ...string) error {
		ran = name == "proot-distro" && strings.Join(args, " ") == "login ubuntu -- true"
		return nil
	}
	service.Deps.EnsureSSHConfigured = func(got paths.Paths) error { configured = got == p; return nil }
	service.Deps.EnsureIfconfig = func(context.Context, io.Writer, func(context.Context, string, ...string) error) error { return nil }
	service.Deps.WakeLock = func() error { locked = true; return nil }
	service.Deps.AcquireLock = func(string) (func(), error) { return func() {}, nil }
	service.Deps.ReadFile = func(string) ([]byte, error) { return nil, os.ErrNotExist }
	service.Deps.ProcessIsMobdeskSSH = func(int, string) bool { return false }
	service.Deps.PortOpen = func(context.Context, int) bool { return false }
	service.Deps.StartSSHD = func(context.Context, string, string) error { started = true; return nil }
	service.Deps.SSHResponds = func(context.Context, int) bool { return true }
	service.Deps.Addresses = func() []string { return []string{"192.168.1.2"} }
	service.Deps.Username = func() string { return "termux" }

	info, err := service.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ran || !configured || !locked || !started {
		t.Fatalf("orquestração incompleta: run=%t config=%t lock=%t start=%t", ran, configured, locked, started)
	}
	if len(info.Addresses) != 1 || info.Addresses[0] != "192.168.1.2" || info.Username != "termux" {
		t.Fatalf("dados de acesso inesperados: %+v", info)
	}
}

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
	service.Deps.Executable = func() (string, error) { return "/bin/mobdesk", nil }
	service.Deps.Abs = func(path string) (string, error) { return path, nil }
	service.Deps.EvalSymlinks = func(path string) (string, error) { return path, nil }

	result, err := service.Setup(context.Background(), SetupOptions{UpgradeSystem: true, AllowPasswordPrompt: true})
	if err != nil {
		t.Fatal(err)
	}
	if !configured || len(result.Phases) != 9 {
		t.Fatalf("setup incompleto: configured=%t phases=%v", configured, result.Phases)
	}
	for _, phase := range result.Phases {
		if _, err := os.Stat(p.SetupPhase(phase)); err != nil {
			t.Fatalf("marker %s ausente: %v", phase, err)
		}
	}
	if _, err := os.Stat(p.SetupDone()); err != nil {
		t.Fatalf("setup.done ausente: %v", err)
	}
	want := []string{"pkg update", "pkg upgrade -y -o Dpkg::Options::=--force-confold", "pkg install -y -o Dpkg::Options::=--force-confold proot-distro openssh net-tools", "proot-distro login ubuntu -- true", "proot-distro install ubuntu", "proot-distro login ubuntu -- mkdir -p /root/workspace /root/.config/mobdesk /root/.local/share/mobdesk", "passwd "}
	if strings.Join(commands, "\n") != strings.Join(want, "\n") {
		t.Fatalf("ordem de comandos inesperada:\n%v", commands)
	}
}

func TestStopRejectsPIDThatIsNotMobdeskSSH(t *testing.T) {
	p := paths.New(t.TempDir(), t.TempDir())
	service := New(p)
	service.Deps.AcquireLock = func(string) (func(), error) { return func() {}, nil }
	service.Deps.ReadFile = func(string) ([]byte, error) { return []byte("42\n"), nil }
	service.Deps.FindProcess = func(int) (Process, error) { return fakeProcess{}, nil }
	service.Deps.ProcessIsMobdeskSSH = func(int, string) bool { return false }
	service.Deps.PortOpen = func(context.Context, int) bool { return true }

	_, err := service.Stop(context.Background())
	if err == nil || !strings.Contains(err.Error(), "não pertence ao servidor SSH do Mobdesk") {
		t.Fatalf("erro inesperado: %v", err)
	}
}

func TestRenderSSHConfigIsDedicatedToMobdesk(t *testing.T) {
	p := paths.New("/home/user", "/termux/usr")
	config := renderSSHConfig(p)
	for _, expected := range []string{"Port 8022", "PidFile /home/user/.local/share/mobdesk/ssh/sshd.pid", "ForceCommand /home/user/.local/share/mobdesk/ssh/mobdesk-ssh-shell", "PasswordAuthentication yes"} {
		if !strings.Contains(config, expected) {
			t.Fatalf("configuração não contém %q:\n%s", expected, config)
		}
	}
}

type fakeProcess struct{}

func (fakeProcess) Signal(os.Signal) error { return errors.New("não deve receber sinal") }
