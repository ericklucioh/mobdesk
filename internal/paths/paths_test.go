package paths

import (
	"path/filepath"
	"testing"
)

func TestLayoutUsesExistingMobdeskPaths(t *testing.T) {
	paths := New("/home/mobdesk", "/termux")
	if paths.DataDir() != "/home/mobdesk/.local/share/mobdesk" {
		t.Fatalf("data dir = %q", paths.DataDir())
	}
	if paths.ConfigDir() != "/home/mobdesk/.config/mobdesk" {
		t.Fatalf("config dir = %q", paths.ConfigDir())
	}
	if paths.SetupPhase("ssh-configured") != "/home/mobdesk/.local/share/mobdesk/state/ssh-configured.done" {
		t.Fatalf("setup phase = %q", paths.SetupPhase("ssh-configured"))
	}
	if paths.SetupDone() != "/home/mobdesk/.local/share/mobdesk/setup.done" || paths.PasswordDone() != "/home/mobdesk/.local/share/mobdesk/password.done" {
		t.Fatalf("unexpected setup markers: %q, %q", paths.SetupDone(), paths.PasswordDone())
	}
	if paths.InstallationsDir() != "/home/mobdesk/.local/share/mobdesk/state/installations" || paths.InstallLogsDir() != "/home/mobdesk/.local/share/mobdesk/logs/install" {
		t.Fatalf("unexpected installation paths: %q, %q", paths.InstallationsDir(), paths.InstallLogsDir())
	}
	if paths.InstallLock() != "/home/mobdesk/.local/share/mobdesk/state/install.lock" {
		t.Fatalf("install lock = %q", paths.InstallLock())
	}
	if paths.SSHConfig() != "/home/mobdesk/.config/mobdesk/ssh/sshd_config" || paths.SSHPID() != "/home/mobdesk/.local/share/mobdesk/ssh/sshd.pid" || paths.SSHLog() != "/home/mobdesk/.local/share/mobdesk/ssh/sshd.log" || paths.SSHWrapper() != "/home/mobdesk/.local/share/mobdesk/ssh/mobdesk-ssh-shell" || paths.SSHLock() != "/home/mobdesk/.local/share/mobdesk/ssh/sshd.lock" {
		t.Fatal("unexpected SSH layout")
	}
	if paths.Launcher() != "/termux/bin/mobdesk" {
		t.Fatalf("launcher = %q", paths.Launcher())
	}
	if paths.UbuntuWorkspace() != "/root/workspace" || paths.UbuntuConfigDir() != "/root/.config/mobdesk" || paths.UbuntuShellConfig() != "/root/.config/mobdesk/bashrc" || paths.UbuntuDataDir() != "/root/.local/share/mobdesk" {
		t.Fatal("unexpected Ubuntu layout")
	}
}

func TestCurrentUsesSafeDefaults(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("PREFIX", "")
	paths := Current()
	if paths.Home != "." {
		t.Fatalf("home = %q, want .", paths.Home)
	}
	if paths.Prefix != defaultPrefix {
		t.Fatalf("prefix = %q", paths.Prefix)
	}
	if paths.DataDir() != filepath.Join(".", ".local", "share", "mobdesk") {
		t.Fatalf("data dir = %q", paths.DataDir())
	}
}

func TestNewPreservesExplicitPaths(t *testing.T) {
	paths := New("", "")
	if paths.Home != "" || paths.Prefix != "" {
		t.Fatalf("New applied implicit defaults: %+v", paths)
	}
}
