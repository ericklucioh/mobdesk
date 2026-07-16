package cobra

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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
		"proot-distro", "openssh", "tmux", "curl", "wget", "git", "termux-services",
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

	ubuntuPackages := []string{
		"ca-certificates", "curl", "wget", "unzip", "zip",
		"build-essential", "pkg-config", "git", "neovim", "tmux",
		"golang", "python3", "python3-pip", "python3-venv",
		"nodejs", "npm", "ripgrep", "fd-find", "fzf", "btop",
	}
	if err := runUbuntu(ctx, "apt-get", "update"); err != nil {
		return err
	}
	if err := runUbuntu(ctx, "apt-get", "upgrade", "-y"); err != nil {
		return err
	}
	args = append([]string{"apt-get", "install", "-y"}, ubuntuPackages...)
	if err := runUbuntu(ctx, args...); err != nil {
		return err
	}
	if err := runUbuntu(ctx, "mkdir", "-p", "/root/workspace", "/root/.config/mobdesk", "/root/.local/share/mobdesk"); err != nil {
		return err
	}
	if err := os.WriteFile(os.ExpandEnv("$HOME/.local/share/mobdesk/setup.done"), []byte("setup concluido\n"), 0o600); err != nil {
		return fmt.Errorf("registrar setup concluído: %w", err)
	}

	fmt.Println("\nSetup concluído.")
	fmt.Println("Ubuntu instalado e ferramentas básicas configuradas.")
	fmt.Println("SSH preparado. Execute: mobdesk start")
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
