package cobra

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/ericklucioh/mobdesk/internal/status"
	"github.com/spf13/cobra"
)

var ErrStatusStrict = errors.New("status incompleto em modo strict")

var (
	statusJSON   bool
	statusStrict bool
)

var statusCmd = newStatusCmd(nil)

func newStatusCmd(state *commandState) *cobra.Command {
	var jsonOutput, strict bool
	cmd := &cobra.Command{Use: "status", Args: localizedNoArgs(state), RunE: func(cmd *cobra.Command, _ []string) error {
		return runStatusOptions(cmd.Context(), jsonOutput, strict, commandLocalizer(state, cmd))
	}}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "")
	cmd.Flags().BoolVar(&strict, "strict", false, "")
	return cmd
}

func runStatus(ctx context.Context) error {
	return runStatusOptions(ctx, statusJSON, statusStrict)
}

func runStatusOptions(ctx context.Context, jsonOutput, strict bool, localizers ...i18n.Localizer) error {
	value := status.Collect(ctx, status.Options{Paths: paths.Current()})
	if jsonOutput {
		if err := status.EncodeJSON(os.Stdout, value); err != nil {
			return fmt.Errorf("%s", localized(localizers, i18n.ErrorOperationFailed, map[string]any{"Detail": err.Error()}, "emitir status JSON: "+err.Error()))
		}
	} else {
		if err := status.RenderText(os.Stdout, value); err != nil {
			return fmt.Errorf("%s", localized(localizers, i18n.ErrorOperationFailed, map[string]any{"Detail": err.Error()}, "emitir status: "+err.Error()))
		}
	}
	if strict && (value.Alerts.Warnings > 0 || value.Alerts.Errors > 0 || value.Alerts.Missing > 0 || value.Alerts.Unknown > 0) {
		return ErrStatusStrict
	}
	return nil
}
