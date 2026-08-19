package cobra

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/ericklucioh/mobdesk/internal/workstation"
	"github.com/spf13/cobra"
)

var migrateJSON bool

var migrateCmd = &cobra.Command{
	Use:   "migrate termux",
	Short: "migrar workspace e ferramentas do Ubuntu para o Termux",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != "termux" {
			return fmt.Errorf("destino de migração não suportado %q; use termux", args[0])
		}
		return runMigrate(cmd.Context(), paths.Current())
	},
}

func init() {
	migrateCmd.Flags().BoolVar(&migrateJSON, "json", false, "emitir apenas JSON válido")
}

func runMigrate(ctx context.Context, p paths.Paths) error {
	var (
		results []install.Result
		err     error
	)
	run := func() error {
		if err := workstation.New(p).MigrateUbuntuWorkspace(ctx); err != nil {
			return err
		}
		results, err = install.MigrateLegacyInstallations(ctx, install.Options{Paths: p})
		return err
	}
	if migrateJSON {
		err = withQuietOutput(run)
		response := operationResult{SchemaVersion: 1, Command: "migrate", Success: err == nil, State: "completed", Message: "Migração para Termux concluída", Migrated: len(results)}
		if err != nil {
			response.State, response.Message = "failed", err.Error()
		}
		if encodeErr := json.NewEncoder(os.Stdout).Encode(response); encodeErr != nil {
			return encodeErr
		}
		return err
	}
	if err := run(); err != nil {
		return err
	}
	fmt.Printf("Migração concluída: workspace em %s; %d ferramenta(s) reinstalada(s) no Termux. Ubuntu foi preservado.\n", p.Workspace(), len(results))
	return nil
}
