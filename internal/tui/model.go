package tui

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ericklucioh/mobdesk/internal/install"
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
	logsScreen
)

// Model coordena navegação e estado compartilhado. Cada tela mantém sua
// renderização em um arquivo próprio; componentes Bubbles ficam reutilizados
// no estado central para que Update e View continuem leves.
type Model struct {
	screen       screen
	status       status.SystemStatus
	version      version.Info
	statusLoaded bool
	width        int
	height       int
	confirmExit  bool
	busy         bool
	message      string
	operation    string
	confirmStop  bool
	dragging     bool
	dragY        int
	pointerDown  bool
	pressMouse   tea.Mouse
	selectedTool int
	focus        int
	history      []screen
	toolsTable   table.Model
	setupActions list.Model
	statusTable  table.Model
	viewport     viewport.Model
	progress     progress.Model
	spinner      spinner.Model
	help         help.Model
}

func New() Model {
	setupDelegate := list.NewDefaultDelegate()
	setupDelegate.ShowDescription = false
	setupDelegate.SetHeight(1)
	setupDelegate.SetSpacing(0)
	setupDelegate.Styles = list.NewDefaultItemStyles(true)
	setupDelegate.Styles.NormalTitle = buttonStyle.Copy().Padding(0, 1)
	setupDelegate.Styles.SelectedTitle = primaryButtonStyle.Copy().Padding(0, 1)
	setupItems := []list.Item{
		setupActionItem{title: "[Enter]  Continuar configuração"},
		setupActionItem{title: "[E]      Executar upgrade completo"},
		setupActionItem{title: "[L]      Ver logs"},
	}
	setupActions := list.New(setupItems, setupDelegate, 40, 5)
	setupActions.SetShowTitle(false)
	setupActions.SetShowFilter(false)
	setupActions.SetShowStatusBar(false)
	setupActions.SetShowPagination(false)
	setupActions.SetShowHelp(false)
	setupActions.DisableQuitKeybindings()

	toolColumns := []table.Column{{Title: "App", Width: 16}, {Title: "Estado", Width: 10}}
	toolsTable := table.New(table.WithColumns(toolColumns), table.WithRows(toolRows(false, 20)))
	toolsTable.SetStyles(toolTableStyles())
	columns := []table.Column{{Title: "Item", Width: 18}, {Title: "Estado", Width: 14}}
	statusTable := table.New(table.WithColumns(columns), table.WithRows(nil))
	tableStyles := table.DefaultStyles()
	tableStyles.Header = tagStyle.Copy().Padding(0, 1)
	tableStyles.Cell = bodyStyle.Copy().Padding(0, 1)
	tableStyles.Selected = cardSelectedStyle.Copy()
	statusTable.SetStyles(tableStyles)

	return Model{
		screen:       homeScreen,
		toolsTable:   toolsTable,
		setupActions: setupActions,
		statusTable:  statusTable,
		viewport:     viewport.New(viewport.WithWidth(40), viewport.WithHeight(18)),
		progress:     progress.New(progress.WithColors(lipgloss.Color(colorPurple), lipgloss.Color(colorLilac)), progress.WithoutPercentage()),
		spinner:      spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(statusColor("starting"))),
		help:         help.New(),
	}
}

type setupActionItem struct{ title string }

func (i setupActionItem) FilterValue() string { return i.title }
func (i setupActionItem) Title() string       { return i.title }
func (i setupActionItem) Description() string { return "" }

func toolTableStyles() table.Styles {
	styles := table.DefaultStyles()
	styles.Header = tagStyle.Copy().Padding(0, 1)
	styles.Cell = bodyStyle.Copy().Padding(0, 1)
	styles.Selected = cardSelectedStyle.Copy()
	return styles
}

func toolRows(loaded bool, width int) []table.Row {
	languages := install.Languages()
	rows := make([]table.Row, 0, len(languages))
	for _, item := range languages {
		state := "disponível"
		if loaded {
			state = "não instalado"
		}
		if width < 60 {
			identity := item.Name
			if width >= 40 {
				identity = item.Name + " · " + item.Package + " · " + item.Executable
			}
			rows = append(rows, table.Row{identity, state})
			continue
		}
		if width < 72 {
			rows = append(rows, table.Row{item.Name, item.Package, state})
			continue
		}
		rows = append(rows, table.Row{item.Name, item.Package, item.Executable, state})
	}
	return rows
}

func (m Model) Run() (tea.Model, error) { return tea.NewProgram(m).Run() }

func (m Model) Init() tea.Cmd {
	return tea.Batch(loadStatus, func() tea.Msg { return m.spinner.Tick() })
}

func loadStatus() tea.Msg {
	return statusMessage{value: status.Collect(context.Background(), status.Options{}), info: version.Current()}
}
