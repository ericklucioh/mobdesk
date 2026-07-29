package cobra

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/ericklucioh/mobdesk/internal/workstation"
	"github.com/spf13/cobra"
)

var startJSON bool
var stopJSON bool

var startCmd = newStartCmd(nil)
var stopCmd = newStopCmd(nil)

func newStartCmd(state *commandState) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{Use: "start", RunE: func(cmd *cobra.Command, _ []string) error {
		localizer := commandLocalizer(state, cmd)
		if jsonOutput {
			return runStartJSON(cmd.Context(), localizer)
		}
		return runStart(cmd.Context(), localizer)
	}}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "")
	return cmd
}

func newStopCmd(state *commandState) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{Use: "stop", RunE: func(cmd *cobra.Command, _ []string) error {
		localizer := commandLocalizer(state, cmd)
		if jsonOutput {
			return runStopJSON(cmd.Context(), localizer)
		}
		return runStop(cmd.Context(), localizer)
	}}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "")
	return cmd
}

func runStart(ctx context.Context, localizers ...i18n.Localizer) error {
	if err := requireTermuxRuntime("mobdesk start", localizers...); err != nil {
		return err
	}
	info, err := workstation.New(paths.Current()).Start(ctx)
	if err != nil {
		return err
	}
	for _, warning := range info.Warnings {
		if _, err := fmt.Fprintln(os.Stdout, localized(localizers, i18n.OutputStartWarning, map[string]any{"Detail": warning}, "Aviso: "+warning)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(os.Stdout, localized(localizers, i18n.OutputStartCompleted, nil, "\nSERVIDOR INICIADO!")); err != nil {
		return err
	}
	if len(info.Addresses) == 0 {
		if _, err := fmt.Fprintln(os.Stdout, localized(localizers, i18n.OutputStartLocalSSH, map[string]any{"Port": workstation.SSHPort, "User": info.Username}, fmt.Sprintf("ACESSO LOCAL VIA SSH\nssh -p %d %s@localhost", workstation.SSHPort, info.Username))); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintln(os.Stdout, localized(localizers, i18n.OutputStartRemoteSSH, nil, "ACESSO REMOTO VIA SSH")); err != nil {
			return err
		}
		for _, address := range info.Addresses {
			if _, err := fmt.Fprintf(os.Stdout, "ssh -p %d %s@%s\n", workstation.SSHPort, info.Username, address); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintln(os.Stdout, localized(localizers, i18n.OutputStartReady, nil, "\nWorkstation pronta. Use mobdesk shell para abrir o Ubuntu.")); err != nil {
		return err
	}
	return nil
}
func runStop(ctx context.Context, localizers ...i18n.Localizer) error {
	if err := requireTermuxRuntime("mobdesk stop", localizers...); err != nil {
		return err
	}
	info, err := workstation.New(paths.Current()).Stop(ctx)
	if err != nil {
		return err
	}
	if info.StaleState {
		_, err = fmt.Fprintln(os.Stdout, localized(localizers, i18n.OutputStopStale, nil, "Servidor SSH já estava parado; estado obsoleto removido."))
	} else if info.AlreadyStopped {
		_, err = fmt.Fprintln(os.Stdout, localized(localizers, i18n.OutputStopAlready, nil, "Servidor SSH já está parado."))
	} else {
		_, err = fmt.Fprintln(os.Stdout, localized(localizers, i18n.OutputStopCompleted, nil, "Servidor SSH parado."))
	}
	return err
}

func runStartJSON(ctx context.Context, localizers ...i18n.Localizer) error {
	var err error
	if runtimeErr := requireTermuxRuntime("mobdesk start", localizers...); runtimeErr != nil {
		err = runtimeErr
	} else {
		if quietErr := withQuietOutput(func() error { _, err = workstation.New(paths.Current()).Start(ctx); return err }); quietErr != nil {
			err = quietErr
		}
	}
	result := operationResult{SchemaVersion: 1, Command: "start", Success: err == nil, State: "running", Message: localized(localizers, i18n.OutputStartCompleted, nil, "Workstation iniciada"), Port: workstation.SSHPort, Addresses: workstation.LocalIPv4Addresses()}
	if err != nil {
		result.State, result.Message = "failed", operationErrorMessage(localizers, err)
	}
	result = decorateResult(result, localizers, i18n.OutputStartCompleted, err)
	if encodeErr := json.NewEncoder(os.Stdout).Encode(result); encodeErr != nil {
		return encodeErr
	}
	return err
}

func runStopJSON(ctx context.Context, localizers ...i18n.Localizer) error {
	var err error
	if runtimeErr := requireTermuxRuntime("mobdesk stop", localizers...); runtimeErr != nil {
		err = runtimeErr
	} else {
		if quietErr := withQuietOutput(func() error { _, err = workstation.New(paths.Current()).Stop(ctx); return err }); quietErr != nil {
			err = quietErr
		}
	}
	result := operationResult{SchemaVersion: 1, Command: "stop", Success: err == nil, State: "stopped", Message: localized(localizers, i18n.OutputStopCompleted, nil, "Workstation parada"), Port: workstation.SSHPort}
	if err != nil {
		result.State, result.Message = "failed", operationErrorMessage(localizers, err)
	}
	result = decorateResult(result, localizers, i18n.OutputStopCompleted, err)
	if encodeErr := json.NewEncoder(os.Stdout).Encode(result); encodeErr != nil {
		return encodeErr
	}
	return err
}
