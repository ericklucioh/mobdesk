// Package paths defines Mobdesk's persistent filesystem layout.
package paths

import (
	"os"
	"path/filepath"
)

const defaultPrefix = "/data/data/com.termux/files/usr"

type Paths struct {
	Home   string
	Prefix string
}

func Current() Paths {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		home = "."
	}
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		prefix = defaultPrefix
	}
	return New(home, prefix)
}

func New(home, prefix string) Paths {
	return Paths{Home: home, Prefix: prefix}
}

func (p Paths) DataDir() string { return filepath.Join(p.Home, ".local", "share", "mobdesk") }

func (p Paths) DataConfigDir() string { return filepath.Join(p.DataDir(), "config") }

func (p Paths) ConfigDir() string { return filepath.Join(p.Home, ".config", "mobdesk") }

func (p Paths) StateDir() string { return filepath.Join(p.DataDir(), "state") }

func (p Paths) SetupPhase(phase string) string { return filepath.Join(p.StateDir(), phase+".done") }

func (p Paths) SetupDone() string { return filepath.Join(p.DataDir(), "setup.done") }

func (p Paths) PasswordDone() string { return filepath.Join(p.DataDir(), "password.done") }

func (p Paths) InstallationsDir() string { return filepath.Join(p.StateDir(), "installations") }

func (p Paths) InstallLogsDir() string { return filepath.Join(p.DataDir(), "logs", "install") }

func (p Paths) SSHConfig() string { return filepath.Join(p.ConfigDir(), "ssh", "sshd_config") }

func (p Paths) SSHRuntimeDir() string { return filepath.Join(p.DataDir(), "ssh") }

func (p Paths) SSHPID() string { return filepath.Join(p.SSHRuntimeDir(), "sshd.pid") }

func (p Paths) SSHLog() string { return filepath.Join(p.SSHRuntimeDir(), "sshd.log") }

func (p Paths) SSHWrapper() string { return filepath.Join(p.SSHRuntimeDir(), "mobdesk-ssh-shell") }

func (p Paths) SSHLock() string { return filepath.Join(p.SSHRuntimeDir(), "sshd.lock") }

func (p Paths) SetupLock() string { return filepath.Join(p.StateDir(), "setup.lock") }

func (p Paths) InstallLock() string { return filepath.Join(p.StateDir(), "install.lock") }

func (p Paths) Launcher() string { return filepath.Join(p.Prefix, "bin", "mobdesk") }

func (p Paths) UbuntuWorkspace() string { return "/root/workspace" }

func (p Paths) UbuntuConfigDir() string { return "/root/.config/mobdesk" }

func (p Paths) UbuntuShellConfig() string { return filepath.Join(p.UbuntuConfigDir(), "bashrc") }

func (p Paths) UbuntuShellLauncher() string { return filepath.Join(p.UbuntuConfigDir(), "shell") }

func (p Paths) UbuntuDataDir() string { return "/root/.local/share/mobdesk" }
