package workstation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ericklucioh/mobdesk/internal/executil"
	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/paths"
)

const (
	defaultTimezone      = "Etc/UTC"
	ubuntuTimezoneScript = `set -eu
zone=$1
test -f "/usr/share/zoneinfo/$zone"
ln -snf "/usr/share/zoneinfo/$zone" /etc/localtime
printf '%s\n' "$zone" > /etc/timezone`
)

type SetupOptions struct {
	UpgradeSystem       bool
	AllowPasswordPrompt bool
}

type SetupResult struct {
	Phases []string
}

func (s Service) Setup(ctx context.Context, options SetupOptions) (result SetupResult, err error) {
	defer func() {
		if err != nil && i18n.ErrorCode(err) == "" {
			err = workstationError("setup workstation", err)
		}
	}()
	result = SetupResult{}
	release, err := s.Deps.AcquireLock(s.Paths.SetupLock())
	if err != nil {
		return result, fmt.Errorf("start setup: %w", err)
	}
	defer release()
	complete := func(phase string) error {
		if err := s.markSetupPhase(phase); err != nil {
			return err
		}
		result.Phases = append(result.Phases, phase)
		return nil
	}
	if !s.setupPhaseDone("directories") {
		if err := s.ensurePrivateDir(filepath.Join(s.Paths.DataDir(), "logs")); err != nil {
			return result, fmt.Errorf("create Mobdesk directories: %w", err)
		}
		if err := s.ensurePrivateDir(s.Paths.DataConfigDir()); err != nil {
			return result, fmt.Errorf("create Mobdesk configuration: %w", err)
		}
		if err := complete("directories"); err != nil {
			return result, err
		}
	}
	if !s.setupPhaseDone("packages-updated") {
		if err := s.run(ctx, "pkg", "update"); err != nil {
			return result, err
		}
		if err := complete("packages-updated"); err != nil {
			return result, err
		}
	}
	if options.UpgradeSystem && !s.setupPhaseDone("system-upgraded") {
		if err := s.run(ctx, "pkg", "upgrade", "-y", "-o", "Dpkg::Options::=--force-confold"); err != nil {
			return result, err
		}
		if err := complete("system-upgraded"); err != nil {
			return result, err
		}
	}
	if !s.setupPhaseDone("packages-installed") {
		if err := s.run(ctx, "pkg", "install", "-y", "-o", "Dpkg::Options::=--force-confold", "proot-distro", "openssh", "net-tools"); err != nil {
			return result, err
		}
		if err := complete("packages-installed"); err != nil {
			return result, err
		}
	}
	if !s.setupPhaseDone("ubuntu-installed") {
		if err := s.ensureUbuntu(ctx); err != nil {
			return result, err
		}
		if err := complete("ubuntu-installed"); err != nil {
			return result, err
		}
	}
	if err := s.configureUbuntuTimezone(ctx); err != nil {
		return result, err
	}
	if !s.setupPhaseDone("workspace-created") {
		if err := s.runUbuntu(ctx, "mkdir", "-p", s.Paths.UbuntuWorkspace(), s.Paths.UbuntuConfigDir(), s.Paths.UbuntuDataDir()); err != nil {
			return result, err
		}
		if err := complete("workspace-created"); err != nil {
			return result, err
		}
	}
	if !s.setupPhaseDone("password-configured") {
		if _, err := s.Deps.Stat(s.Paths.PasswordDone()); err != nil {
			if !os.IsNotExist(err) {
				return result, fmt.Errorf("check SSH password: %w", err)
			}
			if !options.AllowPasswordPrompt {
				return result, fmt.Errorf("SSH password is not configured; run mobdesk setup without --json to configure it")
			}
			if err := s.run(ctx, "passwd"); err != nil {
				return result, fmt.Errorf("configure SSH password: %w", err)
			}
			if err := s.writePrivateFile(s.Paths.PasswordDone(), []byte("senha configurada\n")); err != nil {
				return result, fmt.Errorf("record configured SSH password: %w", err)
			}
		}
		if err := complete("password-configured"); err != nil {
			return result, err
		}
	}
	if !s.setupPhaseDone("ssh-configured") {
		if err := s.Deps.EnsureSSHConfigured(s.Paths); err != nil {
			return result, err
		}
		if err := complete("ssh-configured"); err != nil {
			return result, err
		}
	}
	if !s.setupPhaseDone("shell-configured") {
		zone := s.androidTimezone(ctx)
		dpkgArgs := []string{"dpkg", "--configure", "-a"}
		if !options.AllowPasswordPrompt {
			dpkgArgs = append([]string{"env", "DEBIAN_FRONTEND=noninteractive", "TZ=" + zone}, dpkgArgs...)
		}
		if err := s.runUbuntu(ctx, dpkgArgs...); err != nil {
			return result, fmt.Errorf("repair Ubuntu package state: %w", err)
		}
		if err := s.runUbuntu(ctx, "apt-get", "-y", "update"); err != nil {
			return result, fmt.Errorf("update Ubuntu package lists: %w", err)
		}
		aptArgs := []string{"apt-get", "-o", "DPkg::Lock::Timeout=300", "install", "-y", "bash-completion"}
		if !options.AllowPasswordPrompt {
			aptArgs = append([]string{"env", "DEBIAN_FRONTEND=noninteractive", "TZ=" + zone}, aptArgs...)
		}
		if err := s.runUbuntu(ctx, aptArgs...); err != nil {
			return result, fmt.Errorf("install Bash completion: %w", err)
		}
	}
	// Reconcile generated shell configuration so updates also repair existing setups.
	if err := s.runUbuntu(ctx, "sh", "-ec", renderUbuntuShellConfig(s.Paths)); err != nil {
		return result, fmt.Errorf("configure Ubuntu shell: %w", err)
	}
	if !s.setupPhaseDone("shell-configured") {
		// Rewrite the wrapper even when an older setup already configured SSH.
		if err := s.Deps.EnsureSSHConfigured(s.Paths); err != nil {
			return result, err
		}
		if err := complete("shell-configured"); err != nil {
			return result, err
		}
	}
	if err := s.installLauncher(); err != nil {
		return result, err
	}
	if !s.setupPhaseDone("launcher-installed") {
		if err := complete("launcher-installed"); err != nil {
			return result, err
		}
	}
	if err := s.writePrivateFile(s.Paths.SetupDone(), []byte("setup concluido\n")); err != nil {
		return result, fmt.Errorf("record completed setup: %w", err)
	}
	return result, nil
}

func renderUbuntuShellConfig(p paths.Paths) string {
	return fmt.Sprintf(`set -eu
umask 077
mkdir -p %q
cat > %q <<'EOF'
# Generated by Mobdesk. Interactive Ubuntu shell configuration.
if [ -r /etc/profile ]; then
    . /etc/profile
fi
if [ -r "$HOME/.bashrc" ] && [ "$HOME/.bashrc" != %q ]; then
    . "$HOME/.bashrc"
fi
if [ -r /usr/share/bash-completion/bash_completion ]; then
    . /usr/share/bash-completion/bash_completion
fi
export PATH="$HOME/.local/bin:$PATH"
if command -v javac >/dev/null 2>&1; then
    mobdesk_javac=$(command -v javac)
    mobdesk_javac=$(readlink -f "$mobdesk_javac" 2>/dev/null || printf '%%s' "$mobdesk_javac")
    JAVA_HOME=${mobdesk_javac%%%%/bin/javac}
    export JAVA_HOME
    export PATH="$JAVA_HOME/bin:$PATH"
fi
export SHELL="$HOME/.config/mobdesk/shell"
export CGO_ENABLED=0
PS1='\[\e[35m\]\u@\h\[\e[0m\]:\[\e[36m\]\w\[\e[0m\]\$ '
EOF
chmod 0600 %q
cat > %q <<'EOF'
#!/bin/sh
exec /bin/bash --rcfile %q -i "$@"
EOF
chmod 0700 %q`, p.UbuntuConfigDir(), p.UbuntuShellConfig(), p.UbuntuShellConfig(), p.UbuntuShellConfig(), p.UbuntuShellLauncher(), p.UbuntuShellConfig(), p.UbuntuShellLauncher())
}

func (s Service) setupPhaseDone(phase string) bool {
	_, err := s.Deps.Stat(s.Paths.SetupPhase(phase))
	return err == nil
}

func (s Service) androidTimezone(ctx context.Context) string {
	if s.Deps.AndroidTimezone != nil {
		return s.Deps.AndroidTimezone(ctx)
	}
	return defaultTimezone
}

func (s Service) configureUbuntuTimezone(ctx context.Context) error {
	zone := s.androidTimezone(ctx)
	if !validTimezone(zone) {
		return fmt.Errorf("invalid Android timezone %q", zone)
	}
	if err := s.runUbuntu(ctx, "sh", "-ec", ubuntuTimezoneScript, "--", zone); err != nil {
		return fmt.Errorf("configure Ubuntu timezone %s: %w", zone, err)
	}
	return nil
}

func androidTimezone(ctx context.Context) string {
	for _, executable := range []string{"/system/bin/getprop", "getprop"} {
		command, err := executil.CommandContext(ctx, executable, "persist.sys.timezone")
		if err != nil {
			continue
		}
		output, err := command.Output()
		if err != nil {
			continue
		}
		zone := strings.TrimSpace(string(output))
		if validTimezone(zone) {
			return zone
		}
	}
	return defaultTimezone
}

func validTimezone(zone string) bool {
	if zone == "" || strings.HasPrefix(zone, "/") {
		return false
	}
	for _, part := range strings.Split(zone, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	for index := 0; index < len(zone); index++ {
		character := zone[index]
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || strings.ContainsRune("/_+.-", rune(character)) {
			continue
		}
		return false
	}
	return true
}

func (s Service) markSetupPhase(phase string) error {
	if err := s.ensurePrivateDir(s.Paths.StateDir()); err != nil {
		return fmt.Errorf("create state for phase %s: %w", phase, err)
	}
	if err := s.writePrivateFile(s.Paths.SetupPhase(phase), []byte("concluida\n")); err != nil {
		return fmt.Errorf("record phase %s: %w", phase, err)
	}
	return nil
}

func (s Service) ensurePrivateDir(path string) error {
	if info, err := s.Deps.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refuse private directory that is a symbolic link: %s", path)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("check private directory: %w", err)
	}
	if err := s.Deps.MkdirAll(path, 0o700); err != nil {
		return err
	}
	if err := s.Deps.Chmod(path, 0o700); err != nil {
		return fmt.Errorf("set private permissions: %w", err)
	}
	return nil
}

func (s Service) writePrivateFile(path string, contents []byte) error {
	if info, err := s.Deps.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refuse private file that is a symbolic link: %s", path)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("check private file: %w", err)
	}
	if err := s.Deps.WriteFile(path, contents, 0o600); err != nil {
		return err
	}
	if err := s.Deps.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("set private permissions: %w", err)
	}
	return nil
}

func (s Service) ensureUbuntu(ctx context.Context) error {
	if err := s.run(ctx, "proot-distro", "login", "ubuntu", "--", "true"); err == nil {
		return nil
	}
	return s.run(ctx, "proot-distro", "install", "ubuntu")
}

func (s Service) runUbuntu(ctx context.Context, args ...string) error {
	return s.run(ctx, "proot-distro", append([]string{"login", "ubuntu", "--"}, args...)...)
}

func (s Service) installLauncher() error {
	executable, err := s.Deps.Executable()
	if err != nil {
		return fmt.Errorf("detect Mobdesk executable: %w", err)
	}
	executable, err = s.Deps.Abs(executable)
	if err != nil {
		return fmt.Errorf("resolve Mobdesk executable path: %w", err)
	}
	executable, err = s.Deps.EvalSymlinks(executable)
	if err != nil {
		return fmt.Errorf("resolve Mobdesk executable link: %w", err)
	}
	launcher := s.Paths.Launcher()
	if err := s.Deps.MkdirAll(filepath.Dir(launcher), 0o755); err != nil {
		return fmt.Errorf("create mobdesk command directory: %w", err)
	}
	if info, err := s.Deps.Lstat(launcher); err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("could not create mobdesk command: %s already exists and is not a link", launcher)
		}
		target, err := s.Deps.Readlink(launcher)
		if err != nil {
			return fmt.Errorf("read existing mobdesk command: %w", err)
		}
		if target == executable {
			return nil
		}
		return fmt.Errorf("could not replace mobdesk command: %s points to %s", launcher, target)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check mobdesk command: %w", err)
	}
	if err := s.Deps.Symlink(executable, launcher); err != nil {
		return fmt.Errorf("create mobdesk command: %w", err)
	}
	return nil
}
