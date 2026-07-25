package cobra

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	logstore "github.com/ericklucioh/mobdesk/internal/logs"
	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/spf13/cobra"
)

var (
	logsJSON  bool
	logsName  string
	logsLines int
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "mostrar logs recentes das operações",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runLogs()
	},
}

func init() {
	logsCmd.Flags().BoolVar(&logsJSON, "json", false, "emitir apenas JSON válido")
	logsCmd.Flags().StringVar(&logsName, "name", "", "filtrar pelo nome da instalação")
	logsCmd.Flags().IntVar(&logsLines, "lines", logstore.DefaultLines, "quantidade de linhas recentes por log")
	RootCmd.AddCommand(logsCmd)
}

func runLogs() error {
	snapshot, err := logstore.Read(logstore.Options{Paths: paths.Current(), Name: logsName, Lines: logsLines})
	if err != nil {
		return err
	}
	if logsJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(snapshot)
	}
	if len(snapshot.Logs) == 0 {
		if logsName != "" {
			return fmt.Errorf("nenhum log encontrado para %q", logsName)
		}
		fmt.Println("Nenhum log de instalação registrado.")
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
		fmt.Printf("Log: %s\n", record.LogPath)
		if record.LastError != "" {
			fmt.Printf("Erro: %s\n", record.LastError)
		}
		if record.Missing {
			fmt.Println("Conteúdo: arquivo ausente")
			continue
		}
		if strings.TrimSpace(record.Content) == "" {
			fmt.Println("Conteúdo: vazio")
			continue
		}
		fmt.Println(record.Content)
	}
	return nil
}
