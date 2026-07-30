package tui

import (
	"charm.land/lipgloss/v2"
	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/status"
)

func (m Model) renderSetup() string {
	if !m.canManageHost() {
		return renderPage(m.text(i18n.TUISetupTag, nil), m.text(i18n.TUIHostOnlyTitle, nil), m.text(i18n.TUIHostOnlySetup, nil))
	}
	width := contentWidth(m.width)
	var view string
	view += tagStyle.Render(m.text(i18n.TUISetupTag, nil)) + "\n"
	view += titleStyle.Render(m.text(i18n.TUISetupTitle, nil)) + "\n"
	view += wrapText(m.text(i18n.TUISetupBody, nil), width) + "\n\n"
	m.setupActions.SetSize(max(1, width-4), max(3, min(5, m.height-10)))
	m.setupActions.Select(m.focus)
	view += setupAction(m.setupActions.width, m.setupActions.Index() == 0, m.text(i18n.TUISetupContinue, nil), true) + "\n"
	view += setupAction(m.setupActions.width, m.setupActions.Index() == 1, m.text(i18n.TUISetupUpgrade, nil), false) + "\n"
	view += m.renderSetupStep(width, m.text(i18n.TUISetupDirectories, nil), m.text(i18n.TUISetupDirectoriesDetail, nil), "directories")
	view += m.renderSetupStep(width, m.text(i18n.TUISetupPackages, nil), m.text(i18n.TUISetupPackagesDetail, nil), "packages-installed")
	view += m.renderSetupStep(width, m.text(i18n.TUISetupUbuntu, nil), m.text(i18n.TUISetupUbuntuDetail, nil), "ubuntu-installed")
	view += m.renderSetupStep(width, m.text(i18n.TUISetupWorkspace, nil), m.text(i18n.TUISetupWorkspaceDetail, nil), "workspace-created", "password-configured", "ssh-configured", "shell-configured", "launcher-installed")
	view += "\n"
	view += setupAdvancedLocalized(width, m.focus == 1, m.localizer)
	return view
}

func (m Model) renderSetupStep(width int, title, detail string, phases ...string) string {
	state := setupPhaseState(m.status.Setup.Phases, phases...)
	mark, style := setupStepPresentation(state)
	return setupStep(width, mark, title, detail+" · "+checkStateLabel(state, m.localizer), style)
}

func setupPhaseState(phases map[string]string, required ...string) status.CheckState {
	if len(phases) == 0 {
		return status.CheckUnknown
	}
	for _, phase := range required {
		if phases[phase] != "done" {
			return status.CheckWarning
		}
	}
	return status.CheckOK
}

func setupStepPresentation(state status.CheckState) (string, lipgloss.Style) {
	switch state {
	case status.CheckOK:
		return "✓", stepDoneStyle
	case status.CheckError:
		return "!", stepStyle.Foreground(lipgloss.Color(colorRed)).Bold(true)
	case status.CheckMissing:
		return "?", stepStyle.Foreground(lipgloss.Color(colorYellow))
	case status.CheckUnknown:
		return "?", stepStyle
	default:
		return "○", stepActiveStyle
	}
}

func setupAction(width int, focused bool, label string, primary bool) string {
	style := buttonStyle
	if primary {
		style = primaryButtonStyle
	}
	style = style.Padding(1, 2)
	if focused {
		style = style.BorderForeground(lipgloss.Color(colorLilac)).Bold(true)
		if primary {
			style = style.Background(lipgloss.Color(colorLilac)).Foreground(lipgloss.Color(colorBlack))
		}
	}
	return style.Width(max(1, width)).Render(label)
}

func setupStep(width int, mark, title, detail string, style lipgloss.Style) string {
	text := mark + "  " + title + "\n   " + detail
	return style.Width(max(1, width-2)).Render(text) + "\n"
}

func setupAdvanced(width int, focused bool) string {
	return setupAdvancedLocalized(width, focused, i18n.New(i18n.LocaleENUS))
}

func setupAdvancedLocalized(width int, focused bool, localizer i18n.Localizer) string {
	style := cardStyle.Width(max(1, width-4))
	if focused {
		style = cardSelectedStyle.Width(max(1, width-4))
	}
	text := tagStyle.Render(localizer.Text(i18n.TUISetupAdvanced, nil)) + "\n" + titleStyle.Render(localizer.Text(i18n.TUISetupAdvancedTitle, nil)) + "\n" + wrapText(localizer.Text(i18n.TUISetupAdvancedBody, nil), width-4) + "\n\n" + mutedStyle.Render(localizer.Text(i18n.TUISetupAdvancedHint, nil))
	return style.Render(text)
}
