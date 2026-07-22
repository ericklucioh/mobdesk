package cobra

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/spf13/cobra"
)

var installJSON bool

var installCmd = &cobra.Command{
	Use:   "install <linguagem>",
	Short: "instalar uma linguagem no Ubuntu",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInstall(cmd.Context(), args[0])
	},
}

func init() {
	installCmd.Flags().BoolVar(&installJSON, "json", false, "emitir apenas JSON válido")
	RootCmd.AddCommand(installCmd)
}

func runInstall(ctx context.Context, name string) error {
	result, err := install.Install(ctx, name, install.Options{})
	if err != nil {
		return err
	}
	if installJSON {
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	state := "já estava instalada"
	if result.Changed {
		state = "instalada"
	}
	fmt.Printf("%s %s no Ubuntu (%s): %s\n", strings.Title(result.Language), state, result.Executable, result.Version)
	return nil
}
