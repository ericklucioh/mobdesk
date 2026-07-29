package cobra

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/creack/pty/v2"
	"github.com/ericklucioh/mobdesk/internal/executil"
	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "abrir um shell direto no Ubuntu",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runShell(cmd.Context(), paths.Current())
	},
}

func runShell(ctx context.Context, p paths.Paths) error {
	if err := ensureSetupCompleted(p); err != nil {
		return err
	}
	command, err := executil.CommandContext(ctx, "proot-distro", "login", "ubuntu", "--", "true")
	if err != nil {
		return err
	}
	if err := command.Run(); err != nil {
		return fmt.Errorf("ubuntu não está disponível; execute mobdesk setup: %w", err)
	}
	return runInteractive(ctx, "proot-distro", "login", "ubuntu", "--", "bash", "--rcfile", p.UbuntuShellConfig(), "-i")
}

func ensureSetupCompleted(p paths.Paths) error {
	path := p.SetupDone()
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("setup ainda não foi concluído; execute: mobdesk setup")
		}
		return fmt.Errorf("verificar estado do setup: %w", err)
	}
	return nil
}

func runInteractive(ctx context.Context, name string, args ...string) error {
	fmt.Printf("\n$ %s %s\n", name, strings.Join(args, " "))
	command, err := executil.CommandContext(ctx, name, args...)
	if err != nil {
		return err
	}
	ptmx, err := pty.Start(command)
	if err != nil {
		return fmt.Errorf("iniciar PTY para %q: %w", name, err)
	}
	defer func() { _ = ptmx.Close() }()
	_ = pty.InheritSize(os.Stdin, ptmx)

	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		state, rawErr := term.MakeRaw(fd)
		if rawErr == nil {
			defer func() { _ = term.Restore(fd, state) }()
		}
	}

	inputDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
		close(inputDone)
	}()
	_, copyErr := io.Copy(os.Stdout, ptmx)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if waitErr := command.Wait(); waitErr != nil {
		return fmt.Errorf("comando interativo %q falhou: %w", name, waitErr)
	}
	if copyErr != nil {
		return fmt.Errorf("ler saída interativa %q: %w", name, copyErr)
	}
	select {
	case <-inputDone:
	default:
	}
	return nil
}
