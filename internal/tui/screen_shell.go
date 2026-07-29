package tui

import "charm.land/lipgloss/v2"

func (m Model) renderShell() string {
	width := contentWidth(m.width)
	view := tagStyle.Render("SHELL UBUNTU") + "\n"
	view += titleStyle.Render("Abrir shell") + "\n"
	view += wrapText("A TUI será suspensa enquanto o shell estiver aberto.", width) + "\n\n"
	view += shellAction(width, m.focus == 0, "Abrir shell Ubuntu", "Suspender a TUI e abrir o terminal") + "\n\n"
	view += shellAction(width, m.focus == 1, "Voltar para início", "Retornar à tela principal")
	return view
}

func shellAction(width int, focused bool, title, detail string) string {
	style := buttonStyle.Padding(1, 2)
	if focused {
		style = style.BorderForeground(lipgloss.Color(colorLilac)).Bold(true)
	}
	return style.Width(max(1, width-4)).Render(title + "\n" + mutedStyle.Render(detail))
}
