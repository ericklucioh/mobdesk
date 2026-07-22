package cobra

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "abrir um shell direto no Ubuntu",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runShell(cmd.Context())
	},
}

func runShell(ctx context.Context) error {
	if err := ensureSetupCompleted(); err != nil {
		return err
	}
	if err := runCommand(ctx, "proot-distro", "login", "ubuntu", "--", "true"); err != nil {
		return fmt.Errorf("Ubuntu não está disponível; execute mobdesk setup: %w", err)
	}
	return runInteractive(ctx, "proot-distro", "login", "ubuntu", "--", "bash", "-l")
}

func ensureSetupCompleted() error {
	path := os.ExpandEnv("$HOME/.local/share/mobdesk/setup.done")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("setup ainda não foi concluído; execute: mobdesk setup")
		}
		return fmt.Errorf("verificar estado do setup: %w", err)
	}
	return nil
}
