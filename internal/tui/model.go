package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/ericklucioh/mobdesk/internal/status"
	"github.com/ericklucioh/mobdesk/internal/version"
)

type screen int

const (
	homeScreen screen = iota
	statusScreen
	setupScreen
	toolsScreen
	shellScreen
	systemScreen
)

// Model coordena navegação e estado compartilhado. Cada tela mantém sua
// renderização em um arquivo próprio; componentes Bubbles ficam reutilizados
// no estado central para que Update e View continuem leves.
type Model struct {
	backend           Backend
	screen            screen
	status            status.SystemStatus
	version           version.Info
	statusLoaded      bool
	width             int
	height            int
	confirmExit       bool
	closing           bool
	busy              bool
	operationID       int
	statusID          int
	message           string
	systemMessage     string
	systemState       string
	operation         string
	operationProgress string
	installingTool    string
	confirmStop       bool
	dragging          bool
	dragY             int
	pointerDown       bool
	pressMouse        tea.Mouse
	selectedTool      int
	focus             int
	history           []screen
	toolsList         selector
	setupActions      selector
	statusTable       table.Model
	viewport          viewport.Model
	help              help.Model
}

func New(values ...paths.Paths) Model {
	return newModel(newRealBackend(), tuiPaths(values))
}

// NewWithBackend builds the TUI with an explicit communication backend. It is
// used by the executable mock mode and by focused UI tests.
func NewWithBackend(backend Backend, values ...paths.Paths) Model {
	if backend == nil {
		backend = newRealBackend()
	}
	return newModel(backend, tuiPaths(values))
}

func tuiPaths(values []paths.Paths) paths.Paths {
	if len(values) == 0 {
		return paths.Paths{}
	}
	return values[0]
}

func newModel(backend Backend, p paths.Paths) Model {
	setupActions := selector{count: 2}
	toolsList := selector{count: len(toolEntries(""))}
	columns := statusTableColumns(40)
	statusTable := table.New(table.WithColumns(columns), table.WithRows(nil))
	tableStyles := table.DefaultStyles()
	tableStyles.Header = tagStyle.Padding(0, 1)
	tableStyles.Cell = bodyStyle.Padding(0, 1)
	// Status é uma tabela informativa; não há seleção de linha para destacar.
	tableStyles.Selected = bodyStyle.Padding(0, 1).Foreground(lipgloss.Color(colorLilac)).Bold(true)
	statusTable.SetStyles(tableStyles)

	initialStatus := status.SystemStatus{Installations: status.ReadInstallations(p)}
	return Model{
		backend:      backend,
		screen:       homeScreen,
		status:       initialStatus,
		statusID:     1,
		toolsList:    toolsList,
		setupActions: setupActions,
		statusTable:  statusTable,
		viewport:     viewport.New(viewport.WithWidth(40), viewport.WithHeight(18)),
		help:         help.New(),
	}
}

type toolEntry struct {
	language install.Language
	kind     string
	phrase   string
}

func toolAppLabel(entry toolEntry) string {
	switch entry.language.Name {
	case "python":
		return "python3"
	case "node":
		return "nodejs"
	case "c":
		return "clang"
	case "cpp":
		return "clang++"
	case "lua":
		return "lua5.4"
	default:
		return entry.language.Name
	}
}

func toolEntries(kind string) []toolEntry {
	phrases := map[string]string{
		"go":             "Linguagem compilada",
		"python":         "Scripts e automação",
		"node":           "JavaScript no servidor",
		"c":              "Compilador C",
		"cpp":            "Compilador C++",
		"lua":            "Scripts leves",
		"git":            "Controle de versão",
		"gh":             "GitHub pelo terminal",
		"tmux":           "Sessões persistentes",
		"zellij":         "Multiplexador moderno",
		"micro":          "Editor de terminal",
		"lazygit":        "Git interativo",
		"tree":           "Árvore de diretórios",
		"ttt":            "Editor e IDE de terminal",
		"btop":           "Monitor do sistema",
		"ncdu":           "Uso de disco",
		"inxi":           "Informações do sistema",
		"speedtest-cli":  "Teste de rede",
		"posting":        "Cliente HTTP no terminal",
		"opencode-cli":   "Assistente de IA",
		"codex-cli":      "Assistente de IA",
		"claudecode-cli": "Assistente de IA",
		"leetgo":         "Exercícios LeetCode",
	}
	entries := make([]toolEntry, 0)
	for _, item := range install.Tools() {
		entry := toolEntry{language: item, kind: item.Kind, phrase: phrases[item.Name]}
		if kind == "" || entry.kind == kind {
			entries = append(entries, entry)
		}
	}
	return entries
}

func toolListItems(value status.SystemStatus, installing string) []toolListItem {
	entries := toolEntries("")
	items := make([]toolListItem, 0, len(entries))
	for _, entry := range entries {
		items = append(items, toolListItem{
			entry:      entry,
			installed:  toolInstalled(value, entry),
			installing: entry.language.Name == installing,
		})
	}
	return items
}

func toolInstalled(value status.SystemStatus, entry toolEntry) bool {
	for _, installation := range value.Installations {
		if installation.Kind != "" && installation.Kind != entry.kind {
			continue
		}
		matches := installation.Name == entry.language.Name || installation.Package == entry.language.Package || installation.Executable == entry.language.Executable
		if matches && installation.State == "installed" && installation.LastError == "" {
			return true
		}
	}
	return false
}

func (m Model) Run() (tea.Model, error) { return tea.NewProgram(m).Run() }

func (m Model) Init() tea.Cmd {
	return m.statusCommand(m.statusID)
}
