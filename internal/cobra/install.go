package cobra

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/spf13/cobra"
)

var installJSON bool
var installProgress bool

var installCmd = &cobra.Command{
	Use:   "install <ferramenta>",
	Short: "instalar uma ferramenta no Ubuntu",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInstall(cmd.Context(), args[0])
	},
}

func init() {
	installCmd.Flags().BoolVar(&installJSON, "json", false, "emitir apenas JSON válido")
	installCmd.Flags().BoolVar(&installProgress, "progress", false, "emitir eventos de progresso em JSON")
}

func runInstall(ctx context.Context, name string) error {
	options := install.Options{Paths: paths.Current()}
	if installProgress {
		options.Progress = func(message string) {
			_ = json.NewEncoder(os.Stdout).Encode(installProgressEvent{Event: "progress", Message: message})
		}
	}
	result, err := install.Install(ctx, name, options)
	if installJSON {
		if encodeErr := json.NewEncoder(os.Stdout).Encode(installOperationResult(result, err)); encodeErr != nil {
			return encodeErr
		}
		return err
	}
	if err != nil {
		return err
	}

	state := "já estava instalada"
	if result.Changed {
		state = "instalada"
	}
	fmt.Printf("%s %s no Ubuntu (%s): %s\n", strings.ToUpper(result.Language[:1])+result.Language[1:], state, result.Executable, result.Version)
	return nil
}

type installProgressEvent struct {
	Event   string `json:"event"`
	Message string `json:"message"`
}

func installOperationResult(result install.Result, installErr error) operationResult {
	response := operationResult{
		SchemaVersion: 1,
		Command:       "install",
		Success:       installErr == nil,
		State:         result.State,
		Language:      result.Language,
		Version:       result.Version,
		Message:       "Ferramenta instalada",
	}
	if installErr != nil {
		response.State = "failed"
		response.Message = installErr.Error()
	}
	return response
}
