package cobra

import (
	"fmt"

	"github.com/spf13/cobra"
)

var RootCmd = cobra.Command{
	Use: "mobdesk <text>",
	Short: "Gerenciador de ambiente no android",
	Args: cobra.ArbitraryArgs,
	Run: runMobdesk,
}

var (
	isSetup bool
	isStart bool
)


func Init() {
	RootCmd.Flags().BoolVarP(&isSetup,"setup","se",false,"executar configuração")
	RootCmd.Flags().BoolVarP(&isStart,"start","st",false,"iniciar servidor ssh")
}

func runMobdesk(cmd *cobra.Command, args []string) {
	if isSetup {
		runSetup()
	} else if isStart {
		runStart()
	} else {
		fmt.Println("falto coisa")
	}
}