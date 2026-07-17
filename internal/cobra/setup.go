package cobra

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "configurar o Termux e o Ubuntu",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetup(cmd.Context())
	},
}

func runSetup(ctx context.Context) error {
	if err := os.MkdirAll(os.ExpandEnv("$HOME/.local/share/mobdesk/logs"), 0o700); err != nil {
		return fmt.Errorf("criar diretórios do Mobdesk: %w", err)
	}
	if err := os.MkdirAll(os.ExpandEnv("$HOME/.local/share/mobdesk/config"), 0o700); err != nil {
		return fmt.Errorf("criar configuração do Mobdesk: %w", err)
	}

	termuxPackages := []string{
		// O MVP-1 precisa apenas do runtime Ubuntu e do servidor SSH.
		"proot-distro", "openssh",
	}
	if err := runCommand(ctx, "pkg", "update"); err != nil {
		return err
	}
	if err := runCommand(ctx, "pkg", "upgrade", "-y"); err != nil {
		return err
	}
	args := append([]string{"install", "-y"}, termuxPackages...)
	if err := runCommand(ctx, "pkg", args...); err != nil {
		return err
	}

	if err := ensureUbuntu(ctx); err != nil {
		return err
	}

	if err := runUbuntu(ctx, "mkdir", "-p", "/root/workspace", "/root/.config/mobdesk", "/root/.local/share/mobdesk"); err != nil {
		return err
	}
	if err := ensurePassword(ctx); err != nil {
		return err
	}
	if err := os.WriteFile(os.ExpandEnv("$HOME/.local/share/mobdesk/setup.done"), []byte("setup concluido\n"), 0o600); err != nil {
		return fmt.Errorf("registrar setup concluído: %w", err)
	}
	if err := installLauncher(); err != nil {
		return err
	}

	fmt.Println("\nSetup concluído.")
	fmt.Println("Ubuntu base instalado e pronto para o MVP.")
	fmt.Println("SSH preparado. Execute: mobdesk start")
	return nil
}

func ensurePassword(ctx context.Context) error {
	marker := os.ExpandEnv("$HOME/.local/share/mobdesk/password.done")
	if _, err := os.Stat(marker); err == nil {
		fmt.Println("Senha SSH já configurada.")
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("verificar senha SSH: %w", err)
	}

	fmt.Println("Configure a senha do usuário Termux para acesso via SSH.")
	if err := runCommand(ctx, "passwd"); err != nil {
		return fmt.Errorf("configurar senha SSH: %w", err)
	}
	if err := os.WriteFile(marker, []byte("senha configurada\n"), 0o600); err != nil {
		return fmt.Errorf("registrar senha SSH configurada: %w", err)
	}
	return nil
}

func installLauncher() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("detectar executável do Mobdesk: %w", err)
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return fmt.Errorf("resolver caminho do executável do Mobdesk: %w", err)
	}
	if executable, err = filepath.EvalSymlinks(executable); err != nil {
		return fmt.Errorf("resolver link do executável do Mobdesk: %w", err)
	}

	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		prefix = "/data/data/com.termux/files/usr"
	}
	launcher := filepath.Join(prefix, "bin", "mobdesk")
	if err := os.MkdirAll(filepath.Dir(launcher), 0o755); err != nil {
		return fmt.Errorf("criar diretório do comando mobdesk: %w", err)
	}

	if info, err := os.Lstat(launcher); err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("não foi possível criar o comando mobdesk: %s já existe e não é um link", launcher)
		}
		if err := os.Remove(launcher); err != nil {
			return fmt.Errorf("atualizar comando mobdesk: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("verificar comando mobdesk: %w", err)
	}

	if err := os.Symlink(executable, launcher); err != nil {
		return fmt.Errorf("criar comando mobdesk: %w", err)
	}
	fmt.Printf("Comando disponível globalmente: mobdesk -> %s\n", executable)
	return nil
}

func ensureUbuntu(ctx context.Context) error {
	if err := runCommand(ctx, "proot-distro", "login", "ubuntu", "--", "true"); err == nil {
		fmt.Println("Ubuntu já está instalado.")
		return nil
	}
	fmt.Println("Ubuntu não encontrado; instalando a distribuição persistente...")
	return runCommand(ctx, "proot-distro", "install", "ubuntu")
}

func runUbuntu(ctx context.Context, args ...string) error {
	loginArgs := append([]string{"login", "ubuntu", "--"}, args...)
	return runCommand(ctx, "proot-distro", loginArgs...)
}

func runCommand(ctx context.Context, name string, args ...string) error {
	fmt.Printf("\n$ %s %s\n", name, strings.Join(args, " "))
	command := exec.CommandContext(ctx, name, args...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("comando %q falhou: %w", name, err)
	}
	return nil
}
