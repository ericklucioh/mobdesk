package cobra

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/spf13/cobra"
)

var (
	configApplyJSON      bool
	configApplyProgress  bool
	configRemoveJSON     bool
	configRemoveProgress bool
)

var appConfigCmd = newAppConfigCmd(nil)
var configApplyCmd = appConfigCmd.Commands()[0]
var configRemoveCmd = appConfigCmd.Commands()[1]

func newAppConfigCmd(state *commandState) *cobra.Command {
	var applyJSON, applyProgress, removeJSON, removeProgress bool
	app := &cobra.Command{Use: "config", Args: localizedNoArgs(state)}
	apply := &cobra.Command{Use: "apply <app>", Args: localizedExactArgs(state, 1), RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigOperationOptions(cmd.Context(), args[0], "apply", applyJSON, applyProgress, commandLocalizer(state, cmd))
	}}
	remove := &cobra.Command{Use: "remove <app>", Args: localizedExactArgs(state, 1), RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigOperationOptions(cmd.Context(), args[0], "remove", removeJSON, removeProgress, commandLocalizer(state, cmd))
	}}
	apply.Flags().BoolVar(&applyJSON, "json", false, "")
	apply.Flags().BoolVar(&applyProgress, "progress", false, "")
	remove.Flags().BoolVar(&removeJSON, "json", false, "")
	remove.Flags().BoolVar(&removeProgress, "progress", false, "")
	app.AddCommand(apply, remove)
	return app
}

func runConfigOperation(ctx context.Context, app, action string, jsonOutput, progressOutput bool) error {
	return runConfigOperationOptions(ctx, app, action, jsonOutput, progressOutput)
}

func runConfigOperationOptions(ctx context.Context, app, action string, jsonOutput, progressOutput bool, localizers ...i18n.Localizer) error {
	if runtimeErr := requireTermuxRuntime("mobdesk config "+action, localizers...); runtimeErr != nil {
		return emitConfigResult(install.ConfigOperationResult{App: app, Action: action, State: install.ConfigStateFailed}, runtimeErr, jsonOutput, localizers...)
	}
	var result install.ConfigOperationResult
	var err error
	options := install.Options{Paths: paths.Current()}
	if progressOutput {
		options.Progress = emitInstallProgress
	}
	run := func() error {
		if action == "apply" {
			result, err = install.ApplyConfig(ctx, app, options)
		} else {
			result, err = install.RemoveConfig(ctx, app, options)
		}
		return err
	}
	quietErr := error(nil)
	if progressOutput {
		err = run()
	} else {
		quietErr = withQuietOutput(run)
	}
	if quietErr != nil {
		err = quietErr
	}
	return emitConfigResult(result, err, jsonOutput, localizers...)
}

func emitConfigResult(result install.ConfigOperationResult, operationErr error, jsonOutput bool, localizers ...i18n.Localizer) error {
	if jsonOutput {
		if err := json.NewEncoder(os.Stdout).Encode(configOperationResult(result, operationErr, localizers...)); err != nil {
			return err
		}
		return operationErr
	}
	if operationErr != nil {
		return operationErr
	}
	messageID := i18n.OutputConfigApplied
	if result.Action == "remove" {
		messageID = i18n.OutputConfigRemoved
	}
	fmt.Println(localized(localizers, messageID, nil, result.Message))
	return nil
}

func configOperationResult(result install.ConfigOperationResult, operationErr error, localizers ...i18n.Localizer) operationResult {
	response := operationResult{
		SchemaVersion:   result.SchemaVersion,
		Command:         "config",
		Target:          result.App,
		Action:          result.Action,
		Success:         result.Success && operationErr == nil,
		State:           string(result.State),
		Changed:         result.Changed,
		Message:         localized(localizers, configMessageID(result.Action, result.Success), nil, result.Message),
		ConfigState:     string(result.State),
		Conflicts:       result.Conflicts,
		Paths:           result.Paths,
		StorageEstimate: result.StorageEstimate,
	}
	if response.SchemaVersion == 0 {
		response.SchemaVersion = 1
	}
	if operationErr != nil {
		response.Success = false
		response.State = string(result.State)
		if response.State == "" {
			response.State = string(install.ConfigStateFailed)
		}
		response.ConfigState = response.State
		response.Message = localized(localizers, i18n.OutputConfigFailed, map[string]any{"Detail": operationErr.Error()}, operationErr.Error())
	}
	response = decorateResult(response, localizers, configMessageID(response.Action, response.Success), operationErr)
	return response
}

func configMessageID(action string, success bool) i18n.MessageID {
	if !success {
		return i18n.OutputConfigFailed
	}
	if action == "remove" {
		return i18n.OutputConfigRemoved
	}
	return i18n.OutputConfigApplied
}
