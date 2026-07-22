package cobra

import (
	"github.com/spf13/cobra"
)

var RootCmd = cobra.Command{
	Use:   "mobdesk",
	Short: "Gerenciador de ambiente no android",
}

func init() {
	RootCmd.AddCommand(setupCmd)
	RootCmd.AddCommand(startCmd)
	RootCmd.AddCommand(stopCmd)
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(installCmd)
	RootCmd.AddCommand(shellCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(updateCmd)
}
