package tui

import "charm.land/lipgloss/v2"
import "github.com/ericklucioh/mobdesk/internal/i18n"

func (m Model) renderShell() string {
	width := contentWidth(m.width)
	view := tagStyle.Render(m.text(i18n.TUIShellTag, nil)) + "\n"
	view += titleStyle.Render(m.text(i18n.TUIShellTitle, nil)) + "\n"
	view += wrapText(m.text(i18n.TUIShellBody, nil), width) + "\n\n"
	view += shellAction(width, m.focus == 0, m.text(i18n.TUIShellOpen, nil), m.text(i18n.TUIShellOpenDetail, nil)) + "\n\n"
	view += shellAction(width, m.focus == 1, m.text(i18n.TUIShellBack, nil), m.text(i18n.TUIShellBackDetail, nil))
	return view
}

func shellAction(width int, focused bool, title, detail string) string {
	style := buttonStyle.Padding(1, 2)
	if focused {
		style = style.BorderForeground(lipgloss.Color(colorLilac)).Bold(true)
	}
	return style.Width(max(1, width-4)).Render(title + "\n" + mutedStyle.Render(detail))
}
