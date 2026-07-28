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
}

func runInstall(ctx context.Context, name string) error {
	result, err := install.Install(ctx, name, install.Options{Paths: paths.Current()})
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
