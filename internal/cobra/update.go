package cobra

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/update"
	"github.com/ericklucioh/mobdesk/internal/version"
	"github.com/spf13/cobra"
)

var (
	updateCheck bool
	updateJSON  bool
)

var updateCmd = newUpdateCmd(nil)

func newUpdateCmd(state *commandState) *cobra.Command {
	var check, jsonOutput bool
	cmd := &cobra.Command{Use: "update", Args: localizedNoArgs(state), RunE: func(cmd *cobra.Command, _ []string) error {
		return runUpdateOptions(cmd.Context(), check, jsonOutput, commandLocalizer(state, cmd))
	}}
	cmd.Flags().BoolVar(&check, "check", false, "")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "")
	return cmd
}

func runUpdate(ctx context.Context) error {
	return runUpdateOptions(ctx, updateCheck, updateJSON)
}

func runUpdateOptions(ctx context.Context, checkOnly, jsonOutput bool, localizers ...i18n.Localizer) error {
	info := version.Current()
	options := update.Options{CurrentVersion: info.Version, Channel: info.Channel}
	result := update.Result{CurrentVersion: info.Version, Channel: info.Channel}
	if runtimeErr := requireTermuxRuntime("mobdesk update", localizers...); runtimeErr != nil {
		if jsonOutput {
			response := updateOperationResult(result, runtimeErr, checkOnly, localizers...)
			if encodeErr := json.NewEncoder(os.Stdout).Encode(response); encodeErr != nil {
				return encodeErr
			}
		}
		return runtimeErr
	}
	err := update.Recover(options)
	if err == nil {
		result, err = update.Check(ctx, options)
	}
	if err == nil && !checkOnly && result.Updated {
		result, err = update.Apply(ctx, options)
	}
	if jsonOutput {
		response := updateOperationResult(result, err, checkOnly, localizers...)
		if encodeErr := json.NewEncoder(os.Stdout).Encode(response); encodeErr != nil {
			return encodeErr
		}
		return err
	}
	if err != nil {
		return err
	}
	if !result.Updated {
		fmt.Println(localized(localizers, i18n.OutputUpdateCurrent, map[string]any{"Version": result.CurrentVersion}, fmt.Sprintf("Mobdesk %s já está atualizado.", result.CurrentVersion)))
		return nil
	}
	if checkOnly {
		fmt.Println(localized(localizers, i18n.OutputUpdateAvailable, map[string]any{"Current": result.CurrentVersion, "Latest": result.LatestVersion}, fmt.Sprintf("Atualização disponível: %s → %s", result.CurrentVersion, result.LatestVersion)))
		return nil
	}
	fmt.Println(localized(localizers, i18n.OutputUpdateUpdated, map[string]any{"Current": result.CurrentVersion, "Latest": result.LatestVersion}, fmt.Sprintf("Mobdesk atualizado: %s → %s", result.CurrentVersion, result.LatestVersion)))
	return nil
}

func updateOperationResult(result update.Result, updateErr error, checkOnly bool, localizers ...i18n.Localizer) operationResult {
	response := operationResult{
		SchemaVersion:  1,
		Command:        "update",
		Success:        updateErr == nil,
		CurrentVersion: result.CurrentVersion,
		LatestVersion:  result.LatestVersion,
		Updated:        result.Updated,
	}
	if updateErr != nil {
		response.State = "failed"
		response.Message = operationErrorMessage(localizers, updateErr)
		return decorateResult(response, localizers, i18n.OutputUpdateUpdated, updateErr)
	}
	if !result.Updated {
		response.State = "current"
		response.Message = localized(localizers, i18n.OutputUpdateCurrent, map[string]any{"Version": result.CurrentVersion}, fmt.Sprintf("Mobdesk %s já está atualizado", result.CurrentVersion))
		return decorateResult(response, localizers, i18n.OutputUpdateCurrent, nil)
	}
	if checkOnly {
		response.State = "available"
		response.Message = localized(localizers, i18n.OutputUpdateAvailable, map[string]any{"Current": result.CurrentVersion, "Latest": result.LatestVersion}, fmt.Sprintf("Atualização disponível: %s → %s", result.CurrentVersion, result.LatestVersion))
		return decorateResult(response, localizers, i18n.OutputUpdateAvailable, nil)
	}
	response.State = "updated"
	response.Message = localized(localizers, i18n.OutputUpdateUpdated, map[string]any{"Current": result.CurrentVersion, "Latest": result.LatestVersion}, fmt.Sprintf("Mobdesk atualizado: %s → %s", result.CurrentVersion, result.LatestVersion))
	return decorateResult(response, localizers, i18n.OutputUpdateUpdated, nil)
}
