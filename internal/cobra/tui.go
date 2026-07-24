package cobra

import (
	"github.com/ericklucioh/mobdesk/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "abrir a interface textual do Mobdesk",
	RunE: func(cmd *cobra.Command, _ []string) error {
		_, err := tui.New().Run()
		return err
	},
}
