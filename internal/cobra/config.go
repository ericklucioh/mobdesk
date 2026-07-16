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
	RootCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "iniciar o ambiente e o servidor SSH",
		Run: func(cmd *cobra.Command, args []string) {
			runStart()
		},
	})
}
