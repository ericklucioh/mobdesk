package workstation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type SetupOptions struct {
	UpgradeSystem       bool
	AllowPasswordPrompt bool
}

type SetupResult struct {
	Phases []string
}

func (s Service) Setup(ctx context.Context, options SetupOptions) (SetupResult, error) {
	result := SetupResult{}
	release, err := s.Deps.AcquireLock(s.Paths.SetupLock())
	if err != nil {
		return result, fmt.Errorf("iniciar setup: %w", err)
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
			return result, fmt.Errorf("criar diretórios do Mobdesk: %w", err)
		}
		if err := s.ensurePrivateDir(s.Paths.DataConfigDir()); err != nil {
			return result, fmt.Errorf("criar configuração do Mobdesk: %w", err)
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
				return result, fmt.Errorf("verificar senha SSH: %w", err)
			}
			if !options.AllowPasswordPrompt {
				return result, fmt.Errorf("senha SSH ainda não configurada; execute mobdesk setup sem --json para configurar a senha")
			}
			if err := s.run(ctx, "passwd"); err != nil {
				return result, fmt.Errorf("configurar senha SSH: %w", err)
			}
			if err := s.writePrivateFile(s.Paths.PasswordDone(), []byte("senha configurada\n")); err != nil {
				return result, fmt.Errorf("registrar senha SSH configurada: %w", err)
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
	if err := s.installLauncher(); err != nil {
		return result, err
	}
	if !s.setupPhaseDone("launcher-installed") {
		if err := complete("launcher-installed"); err != nil {
			return result, err
		}
	}
	if err := s.writePrivateFile(s.Paths.SetupDone(), []byte("setup concluido\n")); err != nil {
		return result, fmt.Errorf("registrar setup concluído: %w", err)
	}
	return result, nil
}

func (s Service) setupPhaseDone(phase string) bool {
	_, err := s.Deps.Stat(s.Paths.SetupPhase(phase))
	return err == nil
}

func (s Service) markSetupPhase(phase string) error {
	if err := s.ensurePrivateDir(s.Paths.StateDir()); err != nil {
		return fmt.Errorf("criar estado da etapa %s: %w", phase, err)
	}
	if err := s.writePrivateFile(s.Paths.SetupPhase(phase), []byte("concluida\n")); err != nil {
		return fmt.Errorf("registrar etapa %s: %w", phase, err)
	}
	return nil
}

func (s Service) ensurePrivateDir(path string) error {
	if info, err := s.Deps.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("recusar diretório privado que é link simbólico: %s", path)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("verificar diretório privado: %w", err)
	}
	if err := s.Deps.MkdirAll(path, 0o700); err != nil {
		return err
	}
	if err := s.Deps.Chmod(path, 0o700); err != nil {
		return fmt.Errorf("definir permissões privadas: %w", err)
	}
	return nil
}

func (s Service) writePrivateFile(path string, contents []byte) error {
	if info, err := s.Deps.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("recusar arquivo privado que é link simbólico: %s", path)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("verificar arquivo privado: %w", err)
	}
	if err := s.Deps.WriteFile(path, contents, 0o600); err != nil {
		return err
	}
	if err := s.Deps.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("definir permissões privadas: %w", err)
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
		return fmt.Errorf("detectar executável do Mobdesk: %w", err)
	}
	executable, err = s.Deps.Abs(executable)
	if err != nil {
		return fmt.Errorf("resolver caminho do executável do Mobdesk: %w", err)
	}
	executable, err = s.Deps.EvalSymlinks(executable)
	if err != nil {
		return fmt.Errorf("resolver link do executável do Mobdesk: %w", err)
	}
	launcher := s.Paths.Launcher()
	if err := s.Deps.MkdirAll(filepath.Dir(launcher), 0o755); err != nil {
		return fmt.Errorf("criar diretório do comando mobdesk: %w", err)
	}
	if info, err := s.Deps.Lstat(launcher); err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("não foi possível criar o comando mobdesk: %s já existe e não é um link", launcher)
		}
		target, err := s.Deps.Readlink(launcher)
		if err != nil {
			return fmt.Errorf("ler comando mobdesk existente: %w", err)
		}
		if target == executable {
			return nil
		}
		return fmt.Errorf("não foi possível substituir o comando mobdesk: %s aponta para %s", launcher, target)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("verificar comando mobdesk: %w", err)
	}
	if err := s.Deps.Symlink(executable, launcher); err != nil {
		return fmt.Errorf("criar comando mobdesk: %w", err)
	}
	return nil
}
