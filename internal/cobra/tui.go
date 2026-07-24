package cobra

import (
	"github.com/ericklucioh/mobdesk/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "abrir a interface textual do Mobdesk",
	RunE: func(cmd *cobra.Command, _ []string) error {
		model := tui.New()
		if tuiMock {
			model = tui.NewWithBackend(tui.NewMockBackend(tuiMockScenario))
		}
		_, err := model.Run()
		return err
	},
}

var tuiMock bool
var tuiMockScenario string

func init() {
	tuiCmd.Flags().BoolVar(&tuiMock, "mock", false, "usar backend simulado para testar a TUI")
	tuiCmd.Flags().StringVar(&tuiMockScenario, "mock-scenario", "healthy", "cenário mock: healthy, degraded ou error")
}
