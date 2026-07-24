package tui

import (
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/status"
)

func (m Model) renderHome() string {
	state := "desativado"
	action := "Iniciar"
	if m.status.SSH.Running {
		state, action = "ativado", "Parar"
	}
	stateStyle := homeInactiveStyle
	if m.status.SSH.Running {
		stateStyle = homeActiveStyle
	}
	width := contentWidth(m.width)
	primary := primaryCardStyle.Copy().Width(width)
	cardWidth := width
	if contentColumns(m.width) == 2 {
		cardWidth = (width - 2) / 2
	}
	card := func(index int, icon, title, description string) string {
		marker, style := "  ", cardStyle.Copy()
		if m.focus == index {
			marker, style = "> ", cardSelectedStyle.Copy()
		}
		return style.Width(cardWidth).Render(marker + icon + "  " + titleStyle.Render(title) + "\n" + mutedStyle.Render(description))
	}
	message := ""
	if m.message != "" {
		message = "\n\n" + statusColor("warning").Render(m.message)
	}
	primaryAction := buttonStyle.Render(action)
	primaryText := titleStyle.Render("Workstation SSH") + "\n" + homeStatusLabelStyle.Render("Status: ") + stateStyle.Render(state) + "\n\n" + primaryAction
	access := ""
	if m.status.SSH.Running {
		access = "\n\n" + cardStyle.Copy().Width(max(1, width-4)).Padding(0, 1).Render(
			tagStyle.Render("ACESSO SSH")+"\n"+bodyStyle.Render(m.sshCommand()),
		)
	}
	cards := []string{
		card(1, "◆", "Configurar", "Termux + Ubuntu + SSH"),
		card(2, "◉", "Status", "Ambiente e dispositivo"),
		card(3, "＋", "Apps e linguagens", "Go · Python · Node.js"),
		card(4, "⌁", "Shell Ubuntu", "Abrir terminal"),
		card(5, "◆", "Sistema", "Versão e atualização"),
	}
	return tagStyle.Render("INÍCIO") + "\n" + primary.Render(primaryText) + access + message + "\n\n" + joinCards(cards, m.width)
}

func (m Model) sshCommand() string {
	user := os.Getenv("USER")
	if user == "" {
		user = "android"
	}
	port := m.status.SSH.Port
	if port == 0 {
		port = status.SSHPort
	}
	host := "localhost"
	if len(m.status.Network.Addresses) > 0 && m.status.Network.Addresses[0] != "" {
		host = m.status.Network.Addresses[0]
	}
	return fmt.Sprintf("ssh -p %d %s@%s", port, user, host)
}
