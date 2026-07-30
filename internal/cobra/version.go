package cobra

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/version"
	"github.com/spf13/cobra"
)

var versionJSON bool

var versionCmd = newVersionCmd(nil)

func newVersionCmd(state *commandState) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{Use: "version", Args: localizedNoArgs(state), RunE: func(cmd *cobra.Command, _ []string) error {
		info := version.Current()
		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(info)
		}
		localizer := commandLocalizer(state, cmd)
		fmt.Println(localized([]i18n.Localizer{localizer}, i18n.OutputVersion, map[string]any{"Version": info.Version, "Channel": info.Channel, "OS": info.OS, "Architecture": info.Architecture}))
		return nil
	}}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "")
	return cmd
}
