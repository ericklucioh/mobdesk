package cobra

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/ericklucioh/mobdesk/internal/workstation"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "iniciar o ambiente e o servidor SSH",
	RunE: func(cmd *cobra.Command, args []string) error {
		if startJSON {
			return runStartJSON(cmd.Context())
		}
		return runStart(cmd.Context())
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "parar o servidor SSH",
	RunE: func(cmd *cobra.Command, args []string) error {
		if stopJSON {
			return runStopJSON(cmd.Context())
		}
		return runStop(cmd.Context())
	},
}

var startJSON bool
var stopJSON bool

func init() {
	startCmd.Flags().BoolVar(&startJSON, "json", false, "emitir apenas JSON válido")
	stopCmd.Flags().BoolVar(&stopJSON, "json", false, "emitir apenas JSON válido")
}

func runStart(ctx context.Context) error {
	if err := requireTermuxRuntime("mobdesk start"); err != nil {
		return err
	}
	info, err := workstation.New(paths.Current()).Start(ctx)
	if err != nil {
		return err
	}
	for _, warning := range info.Warnings {
		if _, err := fmt.Fprintf(os.Stdout, "Aviso: %s\n", warning); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(os.Stdout, "\nSERVIDOR INICIADO!"); err != nil {
		return err
	}
	if len(info.Addresses) == 0 {
		if _, err := fmt.Fprintf(os.Stdout, "ACESSO LOCAL VIA SSH\nssh -p %d %s@localhost\n", workstation.SSHPort, info.Username); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintln(os.Stdout, "ACESSO REMOTO VIA SSH"); err != nil {
			return err
		}
		for _, address := range info.Addresses {
			if _, err := fmt.Fprintf(os.Stdout, "ssh -p %d %s@%s\n", workstation.SSHPort, info.Username, address); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintln(os.Stdout, "\nWorkstation pronta. Use mobdesk shell para abrir o Ubuntu."); err != nil {
		return err
	}
	return nil
}
func runStop(ctx context.Context) error {
	if err := requireTermuxRuntime("mobdesk stop"); err != nil {
		return err
	}
	info, err := workstation.New(paths.Current()).Stop(ctx)
	if err != nil {
		return err
	}
	if info.StaleState {
		_, err = fmt.Fprintln(os.Stdout, "Servidor SSH já estava parado; estado obsoleto removido.")
	} else if info.AlreadyStopped {
		_, err = fmt.Fprintln(os.Stdout, "Servidor SSH já está parado.")
	} else {
		_, err = fmt.Fprintln(os.Stdout, "Servidor SSH parado.")
	}
	return err
}

func runStartJSON(ctx context.Context) error {
	var err error
	if runtimeErr := requireTermuxRuntime("mobdesk start"); runtimeErr != nil {
		err = runtimeErr
	} else {
		if quietErr := withQuietOutput(func() error { _, err = workstation.New(paths.Current()).Start(ctx); return err }); quietErr != nil {
			err = quietErr
		}
	}
	result := operationResult{SchemaVersion: 1, Command: "start", Success: err == nil, State: "running", Message: "Workstation iniciada", Port: workstation.SSHPort, Addresses: workstation.LocalIPv4Addresses()}
	if err != nil {
		result.State, result.Message = "failed", err.Error()
	}
	if encodeErr := json.NewEncoder(os.Stdout).Encode(result); encodeErr != nil {
		return encodeErr
	}
	return err
}

func runStopJSON(ctx context.Context) error {
	var err error
	if runtimeErr := requireTermuxRuntime("mobdesk stop"); runtimeErr != nil {
		err = runtimeErr
	} else {
		if quietErr := withQuietOutput(func() error { _, err = workstation.New(paths.Current()).Stop(ctx); return err }); quietErr != nil {
			err = quietErr
		}
	}
	result := operationResult{SchemaVersion: 1, Command: "stop", Success: err == nil, State: "stopped", Message: "Workstation parada", Port: workstation.SSHPort}
	if err != nil {
		result.State, result.Message = "failed", err.Error()
	}
	if encodeErr := json.NewEncoder(os.Stdout).Encode(result); encodeErr != nil {
		return encodeErr
	}
	return err
}
