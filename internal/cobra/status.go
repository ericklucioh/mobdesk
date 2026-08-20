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

var ErrStatusStrict = errors.New("incomplete status in strict mode")

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
	strictFailure := decorateStatusResponse(&value, strict, localizers...)
	if jsonOutput {
		if err := status.EncodeJSON(os.Stdout, value); err != nil {
			return fmt.Errorf("%s", localized(localizers, i18n.ErrorOperationFailed, map[string]any{"Detail": err.Error()}))
		}
	} else {
		if err := status.RenderText(os.Stdout, value, localizers...); err != nil {
			return fmt.Errorf("%s", localized(localizers, i18n.ErrorOperationFailed, map[string]any{"Detail": err.Error()}))
		}
	}
	if strictFailure {
		return ErrStatusStrict
	}
	return nil
}

func decorateStatusResponse(value *status.SystemStatus, strict bool, localizers ...i18n.Localizer) bool {
	value.Command = "status"
	value.Success = true
	value.State = string(value.Overall)
	value.Message = localizedStatusMessage(localizers, value.Overall)
	strictFailure := strict && (value.Alerts.Warnings > 0 || value.Alerts.Errors > 0 || value.Alerts.Missing > 0 || value.Alerts.Unknown > 0)
	if strictFailure {
		value.Success = false
		value.State = "failed"
		value.Message = localized(localizers, i18n.ErrorOperationFailed, map[string]any{"Detail": ErrStatusStrict.Error()})
	}
	return strictFailure
}

func localizedStatusMessage(localizers []i18n.Localizer, overall status.OverallState) string {
	id := i18n.StatusOverallUnknown
	switch overall {
	case status.StateHealthy:
		id = i18n.StatusOverallHealthy
	case status.StateDegraded:
		id = i18n.StatusOverallDegraded
	case status.StateError:
		id = i18n.StatusOverallError
	}
	return localized(localizers, id, nil)
}
