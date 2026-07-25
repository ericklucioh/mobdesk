package cobra

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/ericklucioh/mobdesk/internal/workstation"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use: "setup", Short: "configurar o Termux e o Ubuntu",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := paths.Current()
		if setupJSON {
			return runSetupJSON(cmd.Context(), p)
		}
		return runSetup(cmd.Context(), p)
	},
}

var setupUpgradeSystem bool
var setupJSON bool

func init() {
	setupCmd.Flags().BoolVar(&setupUpgradeSystem, "upgrade-system", false, "atualizar todos os pacotes do Termux antes da instalação")
	setupCmd.Flags().BoolVar(&setupJSON, "json", false, "emitir apenas JSON válido")
}

func runSetup(ctx context.Context, p paths.Paths) error {
	_, err := workstation.New(p).Setup(ctx, workstation.SetupOptions{UpgradeSystem: setupUpgradeSystem, AllowPasswordPrompt: true})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(os.Stdout, "\nSetup concluído."); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(os.Stdout, "Ubuntu base instalado e pronto para o MVP."); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(os.Stdout, "SSH preparado. Execute: mobdesk start"); err != nil {
		return err
	}
	return nil
}

func runSetupJSON(ctx context.Context, p paths.Paths) error {
	var err error
	if quietErr := withQuietOutput(func() error {
		_, err = workstation.New(p).Setup(ctx, workstation.SetupOptions{UpgradeSystem: setupUpgradeSystem})
		return err
	}); quietErr != nil {
		err = quietErr
	}
	result := operationResult{SchemaVersion: 1, Command: "setup", Success: err == nil, State: "completed", Message: "Setup concluído"}
	if err != nil {
		result.State, result.Message = "failed", err.Error()
	}
	if encodeErr := json.NewEncoder(os.Stdout).Encode(result); encodeErr != nil {
		return encodeErr
	}
	return err
}
