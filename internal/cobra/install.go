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

var installJSON bool
var installProgress bool

var installCmd = newInstallCmd(nil)

func newInstallCmd(state *commandState) *cobra.Command {
	var jsonOutput, progressOutput bool
	cmd := &cobra.Command{
		Use:  "install <tool>",
		Args: localizedExactArgs(state, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstallOptions(cmd.Context(), args[0], jsonOutput, progressOutput, commandLocalizer(state, cmd))
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "")
	cmd.Flags().BoolVar(&progressOutput, "progress", false, "")
	return cmd
}

func runInstall(ctx context.Context, name string) error {
	return runInstallOptions(ctx, name, installJSON, installProgress)
}

func runInstallOptions(ctx context.Context, name string, jsonOutput, progressOutput bool, localizers ...i18n.Localizer) error {
	if err := requireTermuxRuntime("mobdesk install", localizers...); err != nil {
		if jsonOutput {
			result := installOperationResult(install.Result{Language: name, State: "failed"}, err, localizers...)
			if encodeErr := json.NewEncoder(os.Stdout).Encode(result); encodeErr != nil {
				return encodeErr
			}
		}
		return err
	}
	options := install.Options{Paths: paths.Current(), Interactive: !jsonOutput && !progressOutput}
	if len(localizers) > 0 {
		options.Localizer = localizers[0]
	}
	if progressOutput {
		options.Progress = func(message string) {
			_ = json.NewEncoder(os.Stdout).Encode(installProgressEvent{Event: "progress", Message: message})
		}
	}
	result, err := install.Install(ctx, name, options)
	if jsonOutput {
		if encodeErr := json.NewEncoder(os.Stdout).Encode(installOperationResult(result, err, localizers...)); encodeErr != nil {
			return encodeErr
		}
		return err
	}
	if err != nil {
		if result.LogPath != "" {
			_, _ = fmt.Fprintln(os.Stderr, localized(localizers, i18n.OutputLogsLabel, map[string]any{"Path": result.LogPath}))
		}
		return err
	}

	id := i18n.OutputInstallAlready
	if result.Changed {
		id = i18n.OutputInstallInstalled
	}
	fmt.Println(localized(localizers, id, map[string]any{"Name": result.Language, "Executable": result.Executable, "Version": result.Version}))
	return nil
}

type installProgressEvent struct {
	Event   string `json:"event"`
	Message string `json:"message"`
}

func installOperationResult(result install.Result, installErr error, localizers ...i18n.Localizer) operationResult {
	response := operationResult{
		SchemaVersion:   1,
		Command:         "install",
		Success:         installErr == nil,
		State:           result.State,
		Target:          result.Language,
		Action:          "install",
		Changed:         result.Changed,
		Language:        result.Language,
		Version:         result.Version,
		LogPath:         result.LogPath,
		Source:          result.Source,
		StorageEstimate: result.StorageEstimate,
		Message:         localized(localizers, i18n.OutputInstallInstalled, nil),
	}
	if installErr != nil {
		response.State = "failed"
		response.Message = operationErrorMessage(localizers, installErr)
	}
	response = decorateResult(response, localizers, i18n.OutputInstallInstalled, installErr)
	return response
}
