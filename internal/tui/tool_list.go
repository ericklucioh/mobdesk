package tui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type toolListItem struct {
	entry     toolEntry
	installed bool
}

func (i toolListItem) FilterValue() string {
	return toolAppLabel(i.entry) + " " + i.entry.phrase
}

type toolListDelegate struct{}

func (toolListDelegate) Height() int                         { return 4 }
func (toolListDelegate) Spacing() int                        { return 0 }
func (toolListDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (toolListDelegate) Render(writer io.Writer, model list.Model, index int, item list.Item) {
	value, ok := item.(toolListItem)
	if !ok {
		return
	}

	width := max(20, model.Width())
	innerWidth := max(1, width-4)
	contentWidth := max(1, innerWidth-4) // borda e padding horizontal
	stateWidth := 10
	if width < 32 {
		stateWidth = 9
	}
	leftWidth := max(1, contentWidth-stateWidth-2)

	appStyle := titleStyle
	if index == model.Index() {
		appStyle = appStyle.Copy().Underline(true)
	}
	app := ansi.Truncate(toolAppLabel(value.entry), leftWidth, "…")
	phrase := ansi.Truncate(value.entry.phrase, leftWidth, "…")
	state := ansi.Truncate(toolDisplayState(value.installed), stateWidth, "…")
	stateStyle := bodyStyle.Copy().Bold(true)
	if value.installed {
		stateStyle = stateStyle.Foreground(lipgloss.Color(colorGreen))
	}
	appView := appStyle.Render(app)
	stateView := stateStyle.Render(state)
	gap := max(2, contentWidth-lipgloss.Width(appView)-lipgloss.Width(stateView))
	firstLine := appView + strings.Repeat(" ", gap) + stateView
	row := firstLine + "\n" + mutedStyle.Render(phrase)
	itemStyle := cardStyle.Copy().Width(innerWidth).Padding(0, 1)
	if index == model.Index() {
		itemStyle = cardSelectedStyle.Copy().Width(innerWidth).Padding(0, 1)
	}
	_, _ = fmt.Fprint(writer, itemStyle.Render(row))
}

func toolDisplayState(installed bool) string {
	if installed {
		return "instalado"
	}
	return "instalar"
}
