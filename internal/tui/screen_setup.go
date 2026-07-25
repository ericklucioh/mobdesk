package tui

import "charm.land/lipgloss/v2"

func (m Model) renderSetup() string {
	if !m.canManageHost() {
		return renderPage("CONFIGURAÇÃO", "Disponível no Termux", "O setup instala e configura componentes do host, como PRoot, SSH e wake-lock. Saia da sessão SSH e execute mobdesk setup no Termux.")
	}
	width := contentWidth(m.width)
	var view string
	view += tagStyle.Render("PRIMEIRO ACESSO") + "\n"
	view += titleStyle.Render("Configurar Mobdesk") + "\n"
	view += wrapText("A configuração é idempotente e pode ser retomada sem apagar seus dados.", width) + "\n\n"
	m.setupActions.SetSize(max(1, width-4), max(3, min(5, m.height-10)))
	m.setupActions.Select(m.focus)
	view += setupAction(m.setupActions.width, m.setupActions.Index() == 0, "[Enter]  Continuar configuração", true) + "\n"
	view += setupAction(m.setupActions.width, m.setupActions.Index() == 1, "[E]      Executar upgrade completo", false) + "\n"
	view += setupStep(width, "✓", "Diretórios do Mobdesk", "config e logs privados criados", stepDoneStyle)
	view += setupStep(width, "✓", "Pacotes Termux", "proot-distro · openssh · net-tools", stepDoneStyle)
	view += setupStep(width, "✓", "Ubuntu persistente", "PRoot ARM64 instalado", stepDoneStyle)
	view += setupStep(width, "●", "Workspace e SSH", "próxima etapa", stepActiveStyle)
	view += "\n"
	view += setupAdvanced(width, m.focus == 1)
	return view
}

func setupAction(width int, focused bool, label string, primary bool) string {
	style := buttonStyle
	if primary {
		style = primaryButtonStyle
	}
	if focused {
		style = cardSelectedStyle.Copy().Padding(0, 1)
		if primary {
			style = primaryButtonStyle.Copy().Bold(true)
		}
	}
	return style.Width(max(1, width-4)).Render(label)
}

func setupStep(width int, mark, title, detail string, style lipgloss.Style) string {
	text := mark + "  " + title + "\n   " + detail
	return style.Copy().Width(max(1, width-2)).Render(text) + "\n"
}

func setupAdvanced(width int, focused bool) string {
	style := cardStyle.Copy().Width(max(1, width-4))
	if focused {
		style = cardSelectedStyle.Copy().Width(max(1, width-4))
	}
	text := tagStyle.Render("OPÇÃO AVANÇADA") + "\n" + titleStyle.Render("Atualizar todo o Termux") + "\n" + wrapText("Equivale a usar setup --upgrade-system.", width-4) + "\n\n" + mutedStyle.Render("Selecione a ação na lista ou pressione E.")
	return style.Render(text)
}
