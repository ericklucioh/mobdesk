package tui

import (
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/status"
)

func (m Model) renderHome() string {
	if !m.canManageHost() {
		width := contentWidth(m.width)
		remote := primaryCardStyle.Width(width).Render(
			titleStyle.Render(m.text(i18n.TUIHomeRemoteTitle, nil)) + "\n" +
				mutedStyle.Render(wrapText(m.text(i18n.TUIHomeRemoteBody, nil), width-4)),
		)
		cards := []string{
			cardStyle.Width(width).Render("◉  " + titleStyle.Render(m.text(i18n.TUIHomeStatusTitle, nil)) + "\n" + mutedStyle.Render(m.text(i18n.TUIHostOnlyHomeStatus, nil))),
			cardStyle.Width(width).Render("⌁  " + titleStyle.Render(m.text(i18n.TUIHomeShellTitle, nil)) + "\n" + mutedStyle.Render(m.text(i18n.TUIHostOnlyHomeShell, nil))),
		}
		message := ""
		if m.message != "" {
			message = "\n\n" + statusColor("warning").Render(m.message)
		}
		return tagStyle.Render(m.text(i18n.TUIHomeTag, nil)) + "\n" + remote + "\n\n" + joinCards(cards, m.width) + message
	}
	state := m.text(i18n.TUIStateStopped, nil)
	action := m.text(i18n.TUIHomeStart, nil)
	if m.status.SSH.Running {
		state, action = m.text(i18n.TUIStateRunning, nil), m.text(i18n.TUIHomeStop, nil)
	}
	stateStyle := homeInactiveStyle
	if m.status.SSH.Running {
		stateStyle = homeActiveStyle
	}
	width := contentWidth(m.width)
	primary := primaryCardStyle.Width(width)
	cardWidth := width
	if contentColumns(m.width) == 2 {
		cardWidth = (width - 2) / 2
	}
	card := func(index int, icon, title, description string) string {
		marker, style := "  ", cardStyle
		if m.focus == index {
			marker, style = "> ", cardSelectedStyle
		}
		return style.Width(cardWidth).Render(marker + icon + "  " + titleStyle.Render(title) + "\n" + mutedStyle.Render(description))
	}
	message := ""
	if m.message != "" {
		message = "\n\n" + statusColor("warning").Render(m.message)
	}
	primaryAction := buttonStyle.Render(action)
	primaryText := titleStyle.Render(m.text(i18n.TUIHomeWorkstationTitle, nil)) + "\n" + homeStatusLabelStyle.Render(m.text(i18n.TUIHomeStatusLabel, nil)) + stateStyle.Render(state) + "\n\n" + primaryAction
	access := ""
	if m.status.SSH.Running {
		access = "\n\n" + cardStyle.Width(max(1, width-4)).Padding(0, 1).Render(
			tagStyle.Render(m.text(i18n.TUIHomeSSHAccess, nil))+"\n"+bodyStyle.Render(m.sshCommand()),
		)
	}
	cards := []string{
		card(1, "◆", m.text(i18n.TUIHomeSetupTitle, nil), m.text(i18n.TUIHomeSetupDetail, nil)),
		card(2, "◉", m.text(i18n.TUIHomeStatusTitle, nil), m.text(i18n.TUIHomeStatusDetail, nil)),
		card(3, "＋", m.text(i18n.TUIHomeAppsTitle, nil), m.text(i18n.TUIHomeAppsDetail, nil)),
		card(4, "⌁", m.text(i18n.TUIHomeShellCardTitle, nil), m.text(i18n.TUIHomeShellCardDetail, nil)),
		card(5, "◆", m.text(i18n.TUIHomeSystemTitle, nil), m.text(i18n.TUIHomeSystemDetail, nil)),
	}
	return tagStyle.Render(m.text(i18n.TUIHomeTag, nil)) + "\n" + primary.Render(primaryText) + access + message + "\n\n" + joinCards(cards, m.width)
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
