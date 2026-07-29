package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/ericklucioh/mobdesk/internal/i18n"
)

func (m Model) renderSystem() string {
	if !m.canManageHost() {
		return renderPage(m.text(i18n.TUISystemTag, nil), m.text(i18n.TUIHostOnlyTitle, nil), m.text(i18n.TUIHostOnlySystem, nil))
	}
	width := contentWidth(m.width)
	versionValue := valueOr(m.version.Version, "dev")
	channelValue := valueOr(m.version.Channel, "dev")
	osValue := valueOr(m.version.OS, m.text(i18n.TUIStatusUnknownArchitecture, nil))
	architectureValue := valueOr(m.version.Architecture, m.text(i18n.TUIStatusUnknownArchitecture, nil))

	check := m.systemAction(0, m.text(i18n.TUISystemCheck, nil), "[V]")
	update := m.systemAction(1, m.text(i18n.TUISystemUpdate, nil), "[A]")
	actions := lipgloss.JoinHorizontal(lipgloss.Top, check, "  ", update)
	if width < 44 {
		actions = lipgloss.JoinVertical(lipgloss.Left, check, update)
	}

	details := []string{
		systemCard(width, m.text(i18n.TUISystemVersion, nil), versionValue),
		systemCard(width, m.text(i18n.TUISystemChannel, nil), channelValue),
		systemCard(width, m.text(i18n.TUISystemPlatform, nil), osValue+"/"+architectureValue),
	}
	advanced := cardStyle.Width(max(1, width-4)).Render(
		tagStyle.Render(m.text(i18n.TUISystemAdvanced, nil)) + "\n" +
			titleStyle.Render(m.text(i18n.TUISystemAdvancedTitle, nil)) + "\n" +
			mutedStyle.Render(wrapText(m.text(i18n.TUISystemAdvancedBody, nil), width-6)),
	)
	feedback := ""
	if m.systemMessage != "" {
		feedback = cardStyle.Width(max(1, width-4)).Render(
			tagStyle.Render(m.text(i18n.TUISystemResult, nil)) + "\n" +
				statusColor(m.systemState).Render("● "+systemResultLabel(m.systemState, m.localizer)) + "\n" +
				bodyStyle.Render(wrapText(m.systemMessage, width-6)),
		)
	}
	back := m.systemAction(2, m.text(i18n.TUIStatusBack, nil), "[Esc]")

	sections := []string{
		tagStyle.Render(m.text(i18n.TUISystemTag, nil)),
		titleStyle.Render(m.text(i18n.TUISystemTitle, nil)),
		mutedStyle.Render(wrapText(m.text(i18n.TUISystemBody, nil), width)),
		cardStyle.Width(max(1, width-4)).Render(
			tagStyle.Render(m.text(i18n.TUISystemUpdate, nil)) + "\n" +
				titleStyle.Render(m.text(i18n.TUISystemTitle, nil)+" "+versionValue) + "\n" +
				mutedStyle.Render(m.text(i18n.TUISystemUpdateHint, nil)) + "\n\n" +
				actions,
		),
	}
	if feedback != "" {
		sections = append(sections, feedback)
	}
	sections = append(sections,
		tagStyle.Render(m.text(i18n.TUISystemVersion, nil)),
		joinCards(details, width),
		advanced,
		back,
	)
	return strings.Join(sections, "\n\n")
}

func systemResultLabel(state string, localizers ...i18n.Localizer) string {
	localizer := i18n.New(i18n.LocaleENUS)
	if len(localizers) > 0 {
		localizer = localizers[0]
	}
	switch state {
	case "current":
		return localizer.Text(i18n.TUISystemCurrent, nil)
	case "available":
		return localizer.Text(i18n.TUISystemAvailable, nil)
	case "updated":
		return localizer.Text(i18n.TUISystemUpdated, nil)
	default:
		return localizer.Text(i18n.TUISystemFailed, nil)
	}
}

func (m Model) systemAction(index int, label, shortcut string) string {
	style := buttonStyle
	if m.focus == index {
		style = primaryButtonStyle
	}
	return style.Render(shortcut + " " + label)
}

func systemCard(width int, label, value string) string {
	cardWidth := contentWidth(width)
	if contentColumns(width) == 2 {
		cardWidth = (cardWidth - 2) / 2
	}
	return cardStyle.Width(cardWidth).Render(tagStyle.Render(label) + "\n" + bodyStyle.Render(value))
}
