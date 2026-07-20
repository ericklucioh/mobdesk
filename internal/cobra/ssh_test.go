package cobra

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderSSHConfigIsDedicatedToMobdesk(t *testing.T) {
	paths := sshPaths{
		prefix:  "/termux/usr",
		config:  "/home/user/.config/mobdesk/ssh/sshd_config",
		pid:     "/home/user/.local/share/mobdesk/ssh/sshd.pid",
		wrapper: "/home/user/.local/share/mobdesk/ssh/mobdesk-ssh-shell",
	}

	config := renderSSHConfig(paths)
	for _, expected := range []string{
		"Port 8022",
		"PidFile /home/user/.local/share/mobdesk/ssh/sshd.pid",
		"ForceCommand /home/user/.local/share/mobdesk/ssh/mobdesk-ssh-shell",
		"PasswordAuthentication yes",
	} {
		if !strings.Contains(config, expected) {
			t.Fatalf("configuração não contém %q:\n%s", expected, config)
		}
	}
	if strings.Contains(config, "PREFIX/etc/ssh/sshd_config") {
		t.Fatal("a configuração dedicada não deve incluir o sshd_config global")
	}
}

func TestProcessIsMobdeskSSHRejectsUnrelatedProcess(t *testing.T) {
	if processIsMobdeskSSH(os.Getpid(), filepath.Join(t.TempDir(), "sshd_config")) {
		t.Fatal("o processo de teste não deve ser identificado como sshd do Mobdesk")
	}
}
