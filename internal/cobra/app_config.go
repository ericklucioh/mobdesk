package cobra

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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

var appConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "gerenciar configurações opcionais de apps",
	Args:  cobra.NoArgs,
}

var configApplyCmd = &cobra.Command{
	Use:   "apply <app>",
	Short: "aplicar a configuração Mobdesk de um app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigOperation(cmd.Context(), args[0], "apply", configApplyJSON, configApplyProgress)
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <app>",
	Short: "remover a configuração Mobdesk de um app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigOperation(cmd.Context(), args[0], "remove", configRemoveJSON, configRemoveProgress)
	},
}

func init() {
	configApplyCmd.Flags().BoolVar(&configApplyJSON, "json", false, "emitir apenas JSON válido")
	configApplyCmd.Flags().BoolVar(&configApplyProgress, "progress", false, "emitir eventos de progresso em JSON")
	configRemoveCmd.Flags().BoolVar(&configRemoveJSON, "json", false, "emitir apenas JSON válido")
	configRemoveCmd.Flags().BoolVar(&configRemoveProgress, "progress", false, "emitir eventos de progresso em JSON")
	appConfigCmd.AddCommand(configApplyCmd, configRemoveCmd)
}

func runConfigOperation(ctx context.Context, app, action string, jsonOutput, progressOutput bool) error {
	if runtimeErr := requireTermuxRuntime("mobdesk config " + action); runtimeErr != nil {
		return emitConfigResult(install.ConfigOperationResult{App: app, Action: action, State: install.ConfigStateFailed}, runtimeErr, jsonOutput)
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
	return emitConfigResult(result, err, jsonOutput)
}

func emitConfigResult(result install.ConfigOperationResult, operationErr error, jsonOutput bool) error {
	if jsonOutput {
		if err := json.NewEncoder(os.Stdout).Encode(configOperationResult(result, operationErr)); err != nil {
			return err
		}
		return operationErr
	}
	if operationErr != nil {
		return operationErr
	}
	fmt.Println(result.Message)
	return nil
}

func configOperationResult(result install.ConfigOperationResult, operationErr error) operationResult {
	response := operationResult{
		SchemaVersion:   result.SchemaVersion,
		Command:         "config",
		Target:          result.App,
		Action:          result.Action,
		Success:         result.Success && operationErr == nil,
		State:           string(result.State),
		Changed:         result.Changed,
		Message:         result.Message,
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
		response.Message = operationErr.Error()
	}
	return response
}
