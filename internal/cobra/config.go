package cobra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type localeContextKey struct{}

type commandState struct {
	environ        func(string) string
	explicitLocale string
}

// RootCmd is kept for callers that use the package directly. New executions
// should use NewRootCmd so environment and Cobra state are isolated per run.
var RootCmd = NewRootCmd()

// NewRootCmd constructs an independent command tree with no shared locale or
// flag state. The environment is read when the tree is executed.
func NewRootCmd() *cobra.Command {
	return newRootCmd(os.Getenv)
}

// NewRootCmdWithEnv is useful for deterministic tests and callers that own
// environment lookup.
func NewRootCmdWithEnv(environ func(string) string) *cobra.Command {
	return newRootCmd(environ)
}

func newRootCmd(environ func(string) string) *cobra.Command {
	state := &commandState{environ: environ}
	root := &cobra.Command{
		Use:               "mobdesk",
		SilenceUsage:      true,
		SilenceErrors:     true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		Args:              localizedRootArgs(state),
		RunE:              func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			locale, err := state.resolve()
			if err != nil {
				return state.validationError(cmd, err)
			}
			cmd.SetContext(context.WithValue(cmd.Context(), localeContextKey{}, i18n.New(locale)))
			return nil
		},
	}
	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		locale := state.fallbackLocalizer()
		message := locale.Text(i18n.ErrorInvalidFlag, map[string]any{"Detail": err.Error()})
		if containsJSONArg(cmd) {
			_ = emitValidationJSON(cmd, locale, message, i18n.ErrorInvalidFlag, "invalid_flag")
		}
		return fmt.Errorf("%s", message)
	})
	root.InitDefaultHelpFlag()
	root.PersistentFlags().StringVar(&state.explicitLocale, "locale", "", "")

	root.AddCommand(
		newSetupCmd(state),
		newStartCmd(state),
		newStopCmd(state),
		newStatusCmd(state),
		newInstallCmd(state),
		newUninstallCmd(state),
		newShellCmd(state),
		newVersionCmd(state),
		newUpdateCmd(state),
		newTUICmd(state),
		newLogsCmd(state),
	)

	localizeHelpTree(root, state)
	return root
}

func (s *commandState) resolve() (i18n.Locale, error) {
	return i18n.Resolve(s.explicitLocale, s.environ)
}

func (s *commandState) fallbackLocalizer() i18n.Localizer {
	locale, err := i18n.Resolve("", s.environ)
	if err != nil {
		locale = i18n.LocaleENUS
	}
	return i18n.New(locale)
}

func (s *commandState) localizer(cmd *cobra.Command) i18n.Localizer {
	if value, ok := cmd.Context().Value(localeContextKey{}).(i18n.Localizer); ok {
		return value
	}
	locale, err := s.resolve()
	if err != nil {
		return s.fallbackLocalizer()
	}
	return i18n.New(locale)
}

func (s *commandState) validationError(cmd *cobra.Command, err error) error {
	locale := s.fallbackLocalizer()
	message := locale.Text(i18n.ErrorInvalidLocale, map[string]any{"Locale": unsupportedLocaleValue(err)})
	if containsJSONArg(cmd) {
		_ = emitValidationJSON(cmd, locale, message, i18n.ErrorInvalidLocale, "invalid_locale")
	}
	return fmt.Errorf("%s", message)
}

func unsupportedLocaleValue(err error) string {
	var unsupported *i18n.UnsupportedLocaleError
	if errors.As(err, &unsupported) {
		return unsupported.Value
	}
	return ""
}

func containsJSONArg(cmd *cobra.Command) bool {
	if flag := cmd.Flags().Lookup("json"); flag != nil && flag.Changed {
		return flag.Value.String() == "true"
	}
	for _, arg := range cmd.Flags().Args() {
		if arg == "--json" {
			return true
		}
	}
	return false
}

func emitValidationJSON(cmd *cobra.Command, localizer i18n.Localizer, message string, messageID i18n.MessageID, errorCode string) error {
	return encodeJSON(operationResult{
		SchemaVersion: 1,
		Command:       cmd.Name(),
		Success:       false,
		State:         "failed",
		Message:       message,
		Locale:        string(localizer.Locale),
		MessageID:     string(messageID),
		ErrorCode:     errorCode,
	})
}

func encodeJSON(value any) error {
	return json.NewEncoder(os.Stdout).Encode(value)
}

func localizeHelpTree(root *cobra.Command, state *commandState) {
	var visit func(*cobra.Command)
	visit = func(cmd *cobra.Command) {
		cmd.SetHelpFunc(func(current *cobra.Command, _ []string) {
			locale, err := state.resolve()
			if err != nil {
				if validationErr := state.validationError(current, err); validationErr != nil {
					_, _ = fmt.Fprintln(current.ErrOrStderr(), validationErr)
				}
				return
			}
			localizer := i18n.New(locale)
			applyHelp(root, localizer)
			_ = printHelp(current, localizer)
		})
		for _, child := range cmd.Commands() {
			visit(child)
		}
	}
	visit(root)
}

type helpSpec struct {
	short   i18n.MessageID
	use     i18n.MessageID
	example i18n.MessageID
}

var helpSpecs = map[string]helpSpec{
	"mobdesk":           {short: i18n.RootShort, example: i18n.RootExample},
	"mobdesk start":     {short: i18n.CommandStartShort, example: i18n.CommandStartExample},
	"mobdesk stop":      {short: i18n.CommandStopShort, example: i18n.CommandStopExample},
	"mobdesk setup":     {short: i18n.CommandSetupShort, example: i18n.CommandSetupExample},
	"mobdesk install":   {short: i18n.CommandInstallShort, use: i18n.CommandInstallUse, example: i18n.CommandInstallExample},
	"mobdesk uninstall": {short: i18n.CommandUninstallShort, use: i18n.CommandUninstallUse, example: i18n.CommandUninstallExample},
	"mobdesk status":    {short: i18n.CommandStatusShort, example: i18n.CommandStatusExample},
	"mobdesk shell":     {short: i18n.CommandShellShort, example: i18n.CommandShellExample},
	"mobdesk version":   {short: i18n.CommandVersionShort, example: i18n.CommandVersionExample},
	"mobdesk update":    {short: i18n.CommandUpdateShort, example: i18n.CommandUpdateExample},
	"mobdesk tui":       {short: i18n.CommandTUIShort, example: i18n.CommandTUIExample},
	"mobdesk logs":      {short: i18n.CommandLogsShort, example: i18n.CommandLogsExample},
}

var flagMessageIDs = map[string]i18n.MessageID{
	"locale":         i18n.FlagLocaleDescription,
	"help":           i18n.FlagHelpDescription,
	"json":           i18n.FlagJSONDescription,
	"progress":       i18n.FlagProgressDescription,
	"upgrade-system": i18n.FlagUpgradeSystemDescription,
	"strict":         i18n.FlagStrictDescription,
	"check":          i18n.FlagCheckDescription,
	"mock":           i18n.FlagMockDescription,
	"mock-scenario":  i18n.FlagMockScenarioDescription,
	"name":           i18n.FlagLogsNameDescription,
	"lines":          i18n.FlagLogsLinesDescription,
}

func applyHelp(root *cobra.Command, localizer i18n.Localizer) {
	var visit func(*cobra.Command)
	visit = func(cmd *cobra.Command) {
		if spec, ok := helpSpecs[cmd.CommandPath()]; ok {
			cmd.Short = localizer.Text(spec.short, nil)
			if spec.use != "" {
				cmd.Use = strings.Fields(cmd.Use)[0] + " " + localizer.Text(spec.use, nil)
			}
			if spec.example != "" {
				cmd.Example = localizer.Text(spec.example, nil)
			}
			if cmd == root {
				cmd.Long = localizer.Text(i18n.RootLong, nil)
			}
		}
		cmd.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
			if id, ok := flagMessageIDs[flag.Name]; ok {
				flag.Usage = localizer.Text(id, nil)
			}
		})
		for _, child := range cmd.Commands() {
			visit(child)
		}
	}
	visit(root)
	if flag := root.PersistentFlags().Lookup("locale"); flag != nil {
		flag.Usage = localizer.Text(i18n.FlagLocaleDescription, nil)
	}
}

func printHelp(cmd *cobra.Command, localizer i18n.Localizer) error {
	out := cmd.OutOrStdout()
	if cmd.Long != "" {
		_, _ = fmt.Fprintln(out, cmd.Long)
	} else if cmd.Short != "" {
		_, _ = fmt.Fprintln(out, cmd.Short)
	}
	_, _ = fmt.Fprintf(out, "\n%s\n  %s\n", localizer.Text(i18n.HelpUsage, nil), helpUseLine(cmd))
	if cmd.Example != "" {
		_, _ = fmt.Fprintf(out, "\n%s\n%s\n", localizer.Text(i18n.HelpExamples, nil), cmd.Example)
	}
	if commands := availableCommands(cmd); len(commands) > 0 {
		_, _ = fmt.Fprintf(out, "\n%s\n", localizer.Text(i18n.HelpCommands, nil))
		for _, child := range commands {
			_, _ = fmt.Fprintf(out, "  %-24s %s\n", helpUseLine(child), child.Short)
		}
	}
	if flags := cmd.NonInheritedFlags().FlagUsages(); flags != "" {
		_, _ = fmt.Fprintf(out, "\n%s\n%s\n", localizer.Text(i18n.HelpFlags, nil), flags)
	}
	if flags := cmd.InheritedFlags().FlagUsages(); flags != "" {
		_, _ = fmt.Fprintf(out, "\n%s\n%s\n", localizer.Text(i18n.HelpGlobalFlags, nil), flags)
	}
	return nil
}

func helpUseLine(cmd *cobra.Command) string {
	fields := strings.Fields(cmd.Use)
	if len(fields) <= 1 {
		return cmd.CommandPath() + " [flags]"
	}
	return cmd.CommandPath() + " " + strings.Join(fields[1:], " ")
}

func availableCommands(cmd *cobra.Command) []*cobra.Command {
	commands := make([]*cobra.Command, 0)
	for _, child := range cmd.Commands() {
		if child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
			commands = append(commands, child)
		}
	}
	return commands
}
