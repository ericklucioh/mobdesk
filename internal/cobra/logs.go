package cobra

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	logstore "github.com/ericklucioh/mobdesk/internal/logs"
	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/spf13/cobra"
)

var (
	logsJSON  bool
	logsName  string
	logsLines int
)

var logsCmd = newLogsCmd(nil)

func newLogsCmd(state *commandState) *cobra.Command {
	var jsonOutput bool
	var name string
	var lines int
	cmd := &cobra.Command{Use: "logs", Args: localizedNoArgs(state), RunE: func(cmd *cobra.Command, _ []string) error {
		return runLogsOptions(jsonOutput, name, lines, commandLocalizer(state, cmd))
	}}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "")
	cmd.Flags().StringVar(&name, "name", "", "")
	cmd.Flags().IntVar(&lines, "lines", logstore.DefaultLines, "")
	return cmd
}

func runLogs() error {
	return runLogsOptions(logsJSON, logsName, logsLines)
}

func runLogsOptions(jsonOutput bool, name string, lines int, localizers ...i18n.Localizer) error {
	if err := requireTermuxRuntime("mobdesk logs", localizers...); err != nil {
		return err
	}
	snapshot, err := logstore.Read(logstore.Options{Paths: paths.Current(), Name: name, Lines: lines})
	if err != nil {
		if len(localizers) > 0 {
			return fmt.Errorf("%s", localizers[0].Error(err))
		}
		return err
	}
	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(snapshot)
	}
	if len(snapshot.Logs) == 0 {
		if name != "" {
			return fmt.Errorf("%s", localized(localizers, i18n.OutputLogsNameEmpty, map[string]any{"Name": name}, fmt.Sprintf("nenhum log encontrado para %q", name)))
		}
		fmt.Println(localized(localizers, i18n.OutputLogsEmpty, nil, "Nenhum log de instalação registrado."))
		return nil
	}
	for index, record := range snapshot.Logs {
		if index > 0 {
			fmt.Println()
		}
		version := record.Version
		if version == "" {
			version = record.State
		}
		fmt.Printf("[%s] %s (%s)\n", record.State, record.Name, version)
		fmt.Println(localized(localizers, i18n.OutputLogsLabel, map[string]any{"Path": record.LogPath}, "Log: "+record.LogPath))
		if record.LastError != "" {
			fmt.Println(localized(localizers, i18n.OutputLogsError, map[string]any{"Detail": record.LastError}, "Erro: "+record.LastError))
		}
		if record.Missing {
			fmt.Println(localized(localizers, i18n.OutputLogsMissing, nil, "Conteúdo: arquivo ausente"))
			continue
		}
		if strings.TrimSpace(record.Content) == "" {
			fmt.Println(localized(localizers, i18n.OutputLogsContentEmpty, nil, "Conteúdo: vazio"))
			continue
		}
		fmt.Println(record.Content)
	}
	return nil
}
