package cobra

import (
	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/ericklucioh/mobdesk/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = newTUICmd(nil)

func newTUICmd(state *commandState) *cobra.Command {
	var mock bool
	scenario := "healthy"
	cmd := &cobra.Command{Use: "tui", RunE: func(cmd *cobra.Command, _ []string) error {
		p := paths.Current()
		localizer := commandLocalizer(state, cmd)
		model := tui.NewWithLocale(localizer.Locale, p)
		if mock {
			model = tui.NewWithBackendLocale(tui.NewMockBackend(scenario), localizer.Locale, p)
		}
		_, err := model.Run()
		return err
	}}
	cmd.Flags().BoolVar(&mock, "mock", false, "")
	cmd.Flags().StringVar(&scenario, "mock-scenario", "healthy", "")
	return cmd
}

var tuiMock bool
var tuiMockScenario string
