package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func (m Model) renderSystem() string {
	if !m.canManageHost() {
		return renderPage("SISTEMA", "Atualizações disponíveis no Termux", "A atualização do Mobdesk altera o binário do host. Saia da sessão SSH e execute mobdesk update no Termux.")
	}
	width := contentWidth(m.width)
	versionValue := valueOr(m.version.Version, "dev")
	channelValue := valueOr(m.version.Channel, "dev")
	osValue := valueOr(m.version.OS, "desconhecido")
	architectureValue := valueOr(m.version.Architecture, "desconhecida")

	check := m.systemAction(0, "Verificar", "[V]")
	update := m.systemAction(1, "Atualizar", "[A]")
	actions := lipgloss.JoinHorizontal(lipgloss.Top, check, "  ", update)
	if width < 44 {
		actions = lipgloss.JoinVertical(lipgloss.Left, check, update)
	}

	details := []string{
		systemCard(width, "VERSÃO", versionValue),
		systemCard(width, "CANAL", channelValue),
		systemCard(width, "PLATAFORMA", osValue+"/"+architectureValue),
	}
	advanced := cardStyle.Width(max(1, width-4)).Render(
		tagStyle.Render("ÁREA AVANÇADA") + "\n" +
			titleStyle.Render("Operações protegidas") + "\n" +
			mutedStyle.Render(wrapText("Ações destrutivas não fazem parte do MVP. O reset do Ubuntu exigirá confirmação explícita.", width-6)),
	)
	back := m.systemAction(2, "Voltar", "[Esc]")

	return strings.Join([]string{
		tagStyle.Render("SISTEMA"),
		titleStyle.Render("Mobdesk"),
		mutedStyle.Render(wrapText("Atualizações, versão e informações do aplicativo.", width)),
		cardStyle.Width(max(1, width-4)).Render(
			tagStyle.Render("ATUALIZAÇÃO") + "\n" +
				titleStyle.Render("Mobdesk "+versionValue) + "\n" +
				mutedStyle.Render("Verifique se há uma versão mais recente.") + "\n\n" +
				actions,
		),
		tagStyle.Render("DETALHES DA VERSÃO"),
		joinCards(details, width),
		advanced,
		back,
	}, "\n\n")
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
