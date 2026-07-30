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
	uninstallJSON     bool
	uninstallProgress bool
)

var uninstallCmd = newUninstallCmd(nil)

func newUninstallCmd(state *commandState) *cobra.Command {
	var jsonOutput, progressOutput bool
	cmd := &cobra.Command{
		Use:  "uninstall <app>",
		Args: localizedExactArgs(state, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstallOptions(cmd.Context(), args[0], jsonOutput, progressOutput, commandLocalizer(state, cmd))
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "")
	cmd.Flags().BoolVar(&progressOutput, "progress", false, "")
	return cmd
}

func runUninstall(ctx context.Context, name string) error {
	return runUninstallOptions(ctx, name, uninstallJSON, uninstallProgress)
}

func runUninstallOptions(ctx context.Context, name string, jsonOutput, progressOutput bool, localizers ...i18n.Localizer) error {
	if runtimeErr := requireTermuxRuntime("mobdesk uninstall", localizers...); runtimeErr != nil {
		return emitUninstallResult(name, install.Result{Language: name, State: "failed"}, runtimeErr, jsonOutput, localizers...)
	}
	var result install.Result
	var err error
	options := install.Options{Paths: paths.Current()}
	if len(localizers) > 0 {
		options.Localizer = localizers[0]
	}
	if progressOutput {
		options.Progress = emitInstallProgress
	}
	run := func() error {
		result, err = install.Uninstall(ctx, name, options)
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
	return emitUninstallResult(name, result, err, jsonOutput, localizers...)
}

func emitInstallProgress(message string) {
	_ = json.NewEncoder(os.Stdout).Encode(installProgressEvent{Event: "progress", Message: message})
}

func emitUninstallResult(name string, result install.Result, operationErr error, jsonOutput bool, localizers ...i18n.Localizer) error {
	if jsonOutput {
		if err := json.NewEncoder(os.Stdout).Encode(uninstallOperationResult(result, name, operationErr, localizers...)); err != nil {
			return err
		}
		return operationErr
	}
	if operationErr != nil {
		return operationErr
	}
	if result.Changed {
		fmt.Println(localized(localizers, i18n.OutputUninstallRemoved, map[string]any{"Name": result.Language}))
	} else {
		fmt.Println(localized(localizers, i18n.OutputUninstallAlready, map[string]any{"Name": result.Language}))
	}
	return nil
}

func uninstallOperationResult(result install.Result, target string, operationErr error, localizers ...i18n.Localizer) operationResult {
	response := operationResult{
		SchemaVersion:   1,
		Command:         "uninstall",
		Target:          target,
		Action:          "uninstall",
		Success:         operationErr == nil,
		State:           result.State,
		Changed:         result.Changed,
		Language:        result.Language,
		Version:         result.Version,
		LogPath:         result.LogPath,
		Source:          result.Source,
		Paths:           result.Paths,
		Conflicts:       result.Conflicts,
		StorageEstimate: result.StorageEstimate,
		Message:         localized(localizers, i18n.OutputUninstallRemoved, map[string]any{"Name": target}),
	}
	if response.State == "" {
		response.State = "failed"
	}
	if operationErr != nil {
		response.State = result.State
		if response.State == "" || response.State == "installed" {
			response.State = "failed"
		}
		response.Message = operationErrorMessage(localizers, operationErr)
	}
	response = decorateResult(response, localizers, i18n.OutputUninstallRemoved, operationErr)
	return response
}
