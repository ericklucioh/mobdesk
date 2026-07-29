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
	uninstallJSON     bool
	uninstallProgress bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <app>",
	Short: "desinstalar um app gerenciado no Ubuntu",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUninstall(cmd.Context(), args[0])
	},
}

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallJSON, "json", false, "emitir apenas JSON válido")
	uninstallCmd.Flags().BoolVar(&uninstallProgress, "progress", false, "emitir eventos de progresso em JSON")
}

func runUninstall(ctx context.Context, name string) error {
	if runtimeErr := requireTermuxRuntime("mobdesk uninstall"); runtimeErr != nil {
		return emitUninstallResult(name, install.Result{Language: name, State: "failed"}, runtimeErr)
	}
	var result install.Result
	var err error
	options := install.Options{Paths: paths.Current()}
	if uninstallProgress {
		options.Progress = emitInstallProgress
	}
	run := func() error {
		result, err = install.Uninstall(ctx, name, options)
		return err
	}
	quietErr := error(nil)
	if uninstallProgress {
		err = run()
	} else {
		quietErr = withQuietOutput(run)
	}
	if quietErr != nil {
		err = quietErr
	}
	return emitUninstallResult(name, result, err)
}

func emitInstallProgress(message string) {
	_ = json.NewEncoder(os.Stdout).Encode(installProgressEvent{Event: "progress", Message: message})
}

func emitUninstallResult(name string, result install.Result, operationErr error) error {
	if uninstallJSON {
		if err := json.NewEncoder(os.Stdout).Encode(uninstallOperationResult(result, name, operationErr)); err != nil {
			return err
		}
		return operationErr
	}
	if operationErr != nil {
		return operationErr
	}
	if result.Changed {
		fmt.Printf("%s desinstalado do Ubuntu.\n", result.Language)
	} else {
		fmt.Printf("%s já estava desinstalado.\n", result.Language)
	}
	return nil
}

func uninstallOperationResult(result install.Result, target string, operationErr error) operationResult {
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
		Message:         "App desinstalado",
	}
	if response.State == "" {
		response.State = "failed"
	}
	if operationErr != nil {
		response.State = result.State
		if response.State == "" || response.State == "installed" {
			response.State = "failed"
		}
		response.Message = operationErr.Error()
	}
	return response
}
