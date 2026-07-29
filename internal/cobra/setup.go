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

var setupUpgradeSystem bool
var setupJSON bool

var setupCmd = newSetupCmd(nil)

func newSetupCmd(state *commandState) *cobra.Command {
	var upgradeSystem, jsonOutput bool
	cmd := &cobra.Command{Use: "setup", RunE: func(cmd *cobra.Command, _ []string) error {
		p := paths.Current()
		localizer := commandLocalizer(state, cmd)
		if jsonOutput {
			return runSetupJSONOptions(cmd.Context(), p, upgradeSystem, localizer)
		}
		return runSetupOptions(cmd.Context(), p, upgradeSystem, localizer)
	}}
	cmd.Flags().BoolVar(&upgradeSystem, "upgrade-system", false, "")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "")
	return cmd
}

func runSetup(ctx context.Context, p paths.Paths, localizers ...i18n.Localizer) error {
	return runSetupOptions(ctx, p, setupUpgradeSystem, localizers...)
}

func runSetupOptions(ctx context.Context, p paths.Paths, upgradeSystem bool, localizers ...i18n.Localizer) error {
	if err := requireTermuxRuntime("mobdesk setup", localizers...); err != nil {
		return err
	}
	_, err := workstation.New(p).Setup(ctx, workstation.SetupOptions{UpgradeSystem: upgradeSystem, AllowPasswordPrompt: true})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(os.Stdout, localized(localizers, i18n.OutputSetupCompleted, nil, "\nSetup concluído.")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(os.Stdout, localized(localizers, i18n.OutputSetupBase, nil, "Ubuntu base instalado e pronto para o MVP.")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(os.Stdout, localized(localizers, i18n.OutputSetupSSH, nil, "SSH preparado. Execute: mobdesk start")); err != nil {
		return err
	}
	return nil
}

func runSetupJSON(ctx context.Context, p paths.Paths, localizers ...i18n.Localizer) error {
	return runSetupJSONOptions(ctx, p, setupUpgradeSystem, localizers...)
}

func runSetupJSONOptions(ctx context.Context, p paths.Paths, upgradeSystem bool, localizers ...i18n.Localizer) error {
	var err error
	if runtimeErr := requireTermuxRuntime("mobdesk setup", localizers...); runtimeErr != nil {
		err = runtimeErr
	} else {
		if quietErr := withQuietOutput(func() error {
			_, err = workstation.New(p).Setup(ctx, workstation.SetupOptions{UpgradeSystem: upgradeSystem})
			return err
		}); quietErr != nil {
			err = quietErr
		}
	}
	result := operationResult{SchemaVersion: 1, Command: "setup", Success: err == nil, State: "completed", Message: localized(localizers, i18n.OutputSetupCompleted, nil, "Setup concluído")}
	if err != nil {
		result.State, result.Message = "failed", operationErrorMessage(localizers, err)
	}
	result = decorateResult(result, localizers, i18n.OutputSetupCompleted, err)
	if encodeErr := json.NewEncoder(os.Stdout).Encode(result); encodeErr != nil {
		return encodeErr
	}
	return err
}
