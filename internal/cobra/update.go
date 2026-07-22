package cobra

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/update"
	"github.com/ericklucioh/mobdesk/internal/version"
	"github.com/spf13/cobra"
)

var (
	updateCheck bool
	updateJSON  bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "verificar e atualizar o Mobdesk",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runUpdate(cmd.Context())
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "apenas verificar se existe atualização")
	updateCmd.Flags().BoolVar(&updateJSON, "json", false, "emitir apenas JSON válido")
}

func runUpdate(ctx context.Context) error {
	info := version.Current()
	options := update.Options{CurrentVersion: info.Version, Channel: info.Channel}
	result, err := update.Check(ctx, options)
	if err != nil {
		return err
	}
	if !updateCheck && result.Updated {
		result, err = update.Apply(ctx, options)
		if err != nil {
			return err
		}
	}
	if updateJSON {
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	if !result.Updated {
		fmt.Printf("Mobdesk %s já está atualizado.\n", result.CurrentVersion)
		return nil
	}
	if updateCheck {
		fmt.Printf("Atualização disponível: %s → %s\n", result.CurrentVersion, result.LatestVersion)
		return nil
	}
	fmt.Printf("Mobdesk atualizado: %s → %s\n", result.CurrentVersion, result.LatestVersion)
	return nil
}
