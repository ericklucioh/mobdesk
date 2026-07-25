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
