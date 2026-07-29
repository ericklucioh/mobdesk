package cobra

import (
	"fmt"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/spf13/cobra"
)

func commandLocalizer(state *commandState, cmd *cobra.Command) i18n.Localizer {
	if state == nil {
		return i18n.New(i18n.LocalePTBR)
	}
	return state.localizer(cmd)
}

func localized(localizers []i18n.Localizer, id i18n.MessageID, data any, fallback string) string {
	if len(localizers) == 0 {
		return fallback
	}
	return localizers[0].Text(id, data)
}

func operationErrorMessage(localizers []i18n.Localizer, err error) string {
	if err == nil {
		return ""
	}
	return localized(localizers, i18n.ErrorOperationFailed, map[string]any{"Detail": err.Error()}, err.Error())
}

func localizedExactArgs(state *commandState, count int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == count {
			return nil
		}
		localizer := commandLocalizer(state, cmd)
		message := localizer.Text(i18n.ValidationExactArgs, map[string]any{"Count": count, "Received": len(args)})
		if containsJSONArg(cmd) {
			_ = emitValidationJSON(cmd, localizer, message, i18n.ErrorInvalidArgs, "invalid_args")
		}
		return fmt.Errorf("%s", message)
	}
}

func localizedNoArgs(state *commandState) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		localizer := commandLocalizer(state, cmd)
		message := localizer.Text(i18n.ValidationNoArgs, map[string]any{"Received": len(args)})
		if containsJSONArg(cmd) {
			_ = emitValidationJSON(cmd, localizer, message, i18n.ErrorInvalidArgs, "invalid_args")
		}
		return fmt.Errorf("%s", message)
	}
}

func localizedRootArgs(state *commandState) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		localizer := commandLocalizer(state, cmd)
		message := localizer.Text(i18n.ErrorUnknownCommand, map[string]any{"Command": args[0]})
		if containsJSONArg(cmd) {
			_ = emitValidationJSON(cmd, localizer, message, i18n.ErrorUnknownCommand, "unknown_command")
		}
		return fmt.Errorf("%s", message)
	}
}
