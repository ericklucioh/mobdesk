package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type toolListItem struct {
	entry      toolEntry
	installed  bool
	installing bool
}

func (i toolListItem) FilterValue() string {
	return toolAppLabel(i.entry) + " " + i.entry.phrase
}

func renderToolItem(value toolListItem, index, selected, width int) string {
	width = max(20, width)
	innerWidth := max(1, width-4)
	contentWidth := max(1, innerWidth-4) // borda e padding horizontal
	stateWidth := 10
	if width < 32 {
		stateWidth = 9
	}
	leftWidth := max(1, contentWidth-stateWidth-2)

	appStyle := titleStyle
	if index == selected {
		appStyle = appStyle.Copy().Underline(true)
	}
	app := ansi.Truncate(toolAppLabel(value.entry), leftWidth, "…")
	phrase := ansi.Truncate(value.entry.phrase, leftWidth, "…")
	state := ansi.Truncate(toolDisplayState(value.installed, value.installing), stateWidth, "…")
	stateStyle := bodyStyle.Copy().Bold(true)
	if value.installed {
		stateStyle = stateStyle.Foreground(lipgloss.Color(colorGreen))
	} else if value.installing {
		stateStyle = stateStyle.Foreground(lipgloss.Color(colorYellow))
	}
	appView := appStyle.Render(app)
	stateView := stateStyle.Render(state)
	gap := max(2, contentWidth-lipgloss.Width(appView)-lipgloss.Width(stateView))
	firstLine := appView + strings.Repeat(" ", gap) + stateView
	row := firstLine + "\n" + mutedStyle.Render(phrase)
	itemStyle := cardStyle.Copy().Width(innerWidth).Padding(0, 1)
	if index == selected {
		itemStyle = cardSelectedStyle.Copy().Width(innerWidth).Padding(0, 1)
	}
	return itemStyle.Render(row)
}

func renderToolItems(items []toolListItem, selected, width, height int) string {
	if len(items) == 0 {
		return ""
	}
	perPage := max(1, height/4)
	start := selected - perPage/2
	start = max(0, min(start, len(items)-perPage))
	end := min(len(items), start+perPage)
	views := make([]string, 0, end-start)
	for index := start; index < end; index++ {
		views = append(views, renderToolItem(items[index], index, selected, width))
	}
	return strings.Join(views, "\n")
}

func toolDisplayState(installed, installing bool) string {
	if installed {
		return "instalado"
	}
	if installing {
		return "Instalando"
	}
	return "instalar"
}
