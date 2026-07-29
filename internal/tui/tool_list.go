package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ericklucioh/mobdesk/internal/i18n"
)

type toolListItem struct {
	entry      toolEntry
	installed  bool
	installing bool
}

func (i toolListItem) FilterValue() string {
	return toolAppLabel(i.entry) + " " + i.entry.profile.Description
}

func renderToolItem(value toolListItem, index, selected, width int) string {
	return renderToolItemLocalized(value, index, selected, width, i18n.New(i18n.LocaleENUS))
}

func renderToolItemLocalized(value toolListItem, index, selected, width int, localizer i18n.Localizer) string {
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
		appStyle = appStyle.Underline(true)
	}
	app := ansi.Truncate(toolAppLabel(value.entry), leftWidth, "…")
	phrase := ansi.Truncate(value.entry.profile.Description, leftWidth, "…")
	state := ansi.Truncate(toolDisplayStateLocalized(value.installed, value.installing, localizer), stateWidth, "…")
	stateStyle := bodyStyle.Bold(true)
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
	itemStyle := cardStyle.Width(innerWidth).Padding(0, 1)
	if index == selected {
		itemStyle = cardSelectedStyle.Width(innerWidth).Padding(0, 1)
	}
	return itemStyle.Render(row)
}

func renderToolItems(items []toolListItem, selected, width, height int) string {
	return renderToolItemsLocalized(items, selected, width, height, i18n.New(i18n.LocaleENUS))
}

func renderToolItemsLocalized(items []toolListItem, selected, width, height int, localizer i18n.Localizer) string {
	if len(items) == 0 {
		return ""
	}
	perPage := max(1, height/4)
	start := selected - perPage/2
	start = max(0, min(start, len(items)-perPage))
	end := min(len(items), start+perPage)
	views := make([]string, 0, end-start)
	for index := start; index < end; index++ {
		views = append(views, renderToolItemLocalized(items[index], index, selected, width, localizer))
	}
	return strings.Join(views, "\n")
}

func toolDisplayState(installed, installing bool) string {
	return toolDisplayStateLocalized(installed, installing, i18n.New(i18n.LocaleENUS))
}

func toolDisplayStateLocalized(installed, installing bool, localizer i18n.Localizer) string {
	if installed {
		return localizer.Text(i18n.TUIToolStateInstalled, nil)
	}
	if installing {
		return localizer.Text(i18n.TUIToolStateInstalling, nil)
	}
	return localizer.Text(i18n.TUIToolStateInstall, nil)
}
