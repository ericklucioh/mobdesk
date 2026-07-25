package cobra

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/version"
	"github.com/spf13/cobra"
)

var versionJSON bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "mostrar a versão do Mobdesk",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		info := version.Current()
		if versionJSON {
			return json.NewEncoder(os.Stdout).Encode(info)
		}
		fmt.Printf("Mobdesk %s (%s) %s/%s\n", info.Version, info.Channel, info.OS, info.Architecture)
		return nil
	},
}

func init() {
	versionCmd.Flags().BoolVar(&versionJSON, "json", false, "emitir apenas JSON válido")
}
