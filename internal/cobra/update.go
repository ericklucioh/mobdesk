package cobra

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/update"
	"github.com/ericklucioh/mobdesk/internal/version"
	"github.com/spf13/cobra"
)

var (
	updateCheck bool
	updateJSON  bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "verificar e atualizar o Mobdesk",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runUpdate(cmd.Context())
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "apenas verificar se existe atualização")
	updateCmd.Flags().BoolVar(&updateJSON, "json", false, "emitir apenas JSON válido")
}

func runUpdate(ctx context.Context) error {
	info := version.Current()
	options := update.Options{CurrentVersion: info.Version, Channel: info.Channel}
	result := update.Result{CurrentVersion: info.Version, Channel: info.Channel}
	err := update.Recover(options)
	if err == nil {
		result, err = update.Check(ctx, options)
	}
	if err == nil && !updateCheck && result.Updated {
		result, err = update.Apply(ctx, options)
	}
	if updateJSON {
		response := updateOperationResult(result, err, updateCheck)
		if encodeErr := json.NewEncoder(os.Stdout).Encode(response); encodeErr != nil {
			return encodeErr
		}
		return err
	}
	if err != nil {
		return err
	}
	if !result.Updated {
		fmt.Printf("Mobdesk %s já está atualizado.\n", result.CurrentVersion)
		return nil
	}
	if updateCheck {
		fmt.Printf("Atualização disponível: %s → %s\n", result.CurrentVersion, result.LatestVersion)
		return nil
	}
	fmt.Printf("Mobdesk atualizado: %s → %s\n", result.CurrentVersion, result.LatestVersion)
	return nil
}

func updateOperationResult(result update.Result, updateErr error, checkOnly bool) operationResult {
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
		response.Message = updateErr.Error()
		return response
	}
	if !result.Updated {
		response.State = "current"
		response.Message = fmt.Sprintf("Mobdesk %s já está atualizado", result.CurrentVersion)
		return response
	}
	if checkOnly {
		response.State = "available"
		response.Message = fmt.Sprintf("Atualização disponível: %s → %s", result.CurrentVersion, result.LatestVersion)
		return response
	}
	response.State = "updated"
	response.Message = fmt.Sprintf("Mobdesk atualizado: %s → %s", result.CurrentVersion, result.LatestVersion)
	return response
}
