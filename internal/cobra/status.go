package cobra

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/status"
	"github.com/spf13/cobra"
)

var ErrStatusStrict = errors.New("status incompleto em modo strict")

var (
	statusJSON   bool
	statusStrict bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "mostrar o estado do ambiente",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runStatus(cmd.Context())
	},
}

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "emitir apenas JSON válido")
	statusCmd.Flags().BoolVar(&statusStrict, "strict", false, "falhar quando houver coleta parcial")
}

func runStatus(ctx context.Context) error {
	value := status.Collect(ctx, status.Options{})
	if statusJSON {
		if err := status.EncodeJSON(os.Stdout, value); err != nil {
			return fmt.Errorf("emitir status JSON: %w", err)
		}
	} else {
		status.RenderText(os.Stdout, value)
	}
	if statusStrict && (value.Alerts.Warnings > 0 || value.Alerts.Errors > 0 || value.Alerts.Missing > 0 || value.Alerts.Unknown > 0) {
		return ErrStatusStrict
	}
	return nil
}
