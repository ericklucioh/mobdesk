package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ericklucioh/mobdesk/internal/i18n"
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

// Model coordinates navigation and shared state. Each screen keeps its
// rendering in its own file; Bubbles components are reused in the central
// state so Update and View remain lightweight.
type Model struct {
	backend           Backend
	localizer         i18n.Localizer
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
	appPopupOpen      bool
	popupAppIndex     int
	popupFocus        int
	popupAction       string
	popupConfirm      bool
	popupMessage      string
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
	return NewWithLocale(i18n.LocaleENUS, values...)
}

// NewWithLocale forwards the selected CLI locale to child Mobdesk commands.
func NewWithLocale(locale i18n.Locale, values ...paths.Paths) Model {
	return newLocalizedModel(newRealBackend(locale), i18n.New(locale), tuiPaths(values))
}

// NewWithBackend builds the TUI with an explicit communication backend. It is
// used by the executable mock mode and by focused UI tests.
func NewWithBackend(backend Backend, values ...paths.Paths) Model {
	return NewWithBackendLocale(backend, i18n.LocaleENUS, values...)
}

func NewWithBackendLocale(backend Backend, locale i18n.Locale, values ...paths.Paths) Model {
	if backend == nil {
		backend = newRealBackend(locale)
	}
	return newLocalizedModel(backend, i18n.New(locale), tuiPaths(values))
}

func tuiPaths(values []paths.Paths) paths.Paths {
	if len(values) == 0 {
		return paths.Paths{}
	}
	return values[0]
}

func newModel(backend Backend, p paths.Paths) Model {
	return newLocalizedModel(backend, i18n.New(i18n.LocaleENUS), p)
}

func newLocalizedModel(backend Backend, localizer i18n.Localizer, p paths.Paths) Model {
	setupActions := selector{count: 2}
	toolsList := selector{count: len(toolEntriesLocalized("", localizer))}
	columns := statusTableColumns(40, localizer)
	statusTable := table.New(table.WithColumns(columns), table.WithRows(nil))
	tableStyles := table.DefaultStyles()
	tableStyles.Header = tagStyle.Padding(0, 1)
	tableStyles.Cell = bodyStyle.Padding(0, 1)
	// Status is an informational table; it has no selectable row to highlight.
	tableStyles.Selected = bodyStyle.Padding(0, 1).Foreground(lipgloss.Color(colorLilac)).Bold(true)
	statusTable.SetStyles(tableStyles)

	initialStatus := status.SystemStatus{Installations: status.ReadInstallations(p)}
	return Model{
		backend:      backend,
		localizer:    localizer,
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
	profile install.AppProfile
	kind    string
}

func toolAppLabel(entry toolEntry) string {
	switch entry.profile.Name {
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
		return entry.profile.Name
	}
}

func toolEntries(kind string) []toolEntry {
	return toolEntriesLocalized(kind, i18n.New(i18n.LocaleENUS))
}

func toolEntriesLocalized(kind string, localizer i18n.Localizer) []toolEntry {
	entries := make([]toolEntry, 0)
	for _, item := range install.Tools(localizer) {
		entry := toolEntry{profile: item, kind: item.Kind}
		if kind == "" || entry.kind == kind {
			entries = append(entries, entry)
		}
	}
	return entries
}

func toolListItems(value status.SystemStatus, installing string) []toolListItem {
	return toolListItemsLocalized(value, installing, i18n.New(i18n.LocaleENUS))
}

func toolListItemsLocalized(value status.SystemStatus, installing string, localizer i18n.Localizer) []toolListItem {
	entries := toolEntriesLocalized("", localizer)
	items := make([]toolListItem, 0, len(entries))
	for _, entry := range entries {
		items = append(items, toolListItem{
			entry:      entry,
			installed:  toolInstalled(value, entry),
			partial:    toolPartial(value, entry),
			installing: entry.profile.Name == installing,
		})
	}
	return items
}

func toolPartial(value status.SystemStatus, entry toolEntry) bool {
	for _, installation := range value.Installations {
		matches := installation.Name == entry.profile.Name || installation.Package == entry.profile.Package || installation.Executable == entry.profile.Executable
		if matches && installation.State == "partial" {
			return true
		}
	}
	return false
}

func toolInstalled(value status.SystemStatus, entry toolEntry) bool {
	for _, installation := range value.Installations {
		if installation.Kind != "" && installation.Kind != entry.kind {
			continue
		}
		matches := installation.Name == entry.profile.Name || installation.Package == entry.profile.Package || installation.Executable == entry.profile.Executable
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
