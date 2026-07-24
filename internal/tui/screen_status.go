package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/ericklucioh/mobdesk/internal/status"
)

func (m Model) renderStatus() string {
	if !m.statusLoaded {
		return renderPage("STATUS", "Carregando", "Coletando estado do ambiente...")
	}

	generated := ""
	if !m.status.GeneratedAt.IsZero() {
		generated = " · " + m.status.GeneratedAt.Format("15:04")
	}

	hostTitle := "TERMUX"
	hostDetail := fmt.Sprintf("%s · %s", valueOr(m.status.Host.OS, "host Android"), valueOr(m.status.Host.Architecture, "arquitetura desconhecida"))
	ubuntuDetail := fmt.Sprintf("PRoot · workspace %s", yesNo(m.status.Ubuntu.Workspace))
	sshTitle := "SSH"
	if !m.status.Host.Termux {
		hostTitle = "RUNTIME"
		hostDetail = fmt.Sprintf("%s · fora do Termux", valueOr(m.status.Host.OS, "Linux"))
		ubuntuDetail = fmt.Sprintf("workspace %s · sessão remota", yesNo(m.status.Ubuntu.Workspace))
		sshTitle = "SSH HOST"
	}
	cards := []string{
		statusCard(m.width, hostTitle, m.status.Host.State, hostDetail),
		statusCard(m.width, "UBUNTU", m.status.Ubuntu.State, ubuntuDetail),
		statusCard(m.width, sshTitle, sshState(m.status.SSH), sshDetail(m.status.SSH)),
		statusCard(m.width, "RECURSOS", m.status.Storage.State, fmt.Sprintf("%s livres · bateria %s", formatBytes(m.status.Storage.DeviceFree), batterySummary(m.status.Battery))),
	}

	network := statusNetworkSummary(m.width, m.status.Network)
	installations := fmt.Sprintf("%d instalação(ões)", len(m.status.Installations))
	alerts := fmt.Sprintf("%d OK · %d avisos · %d erros", m.status.Alerts.OK, m.status.Alerts.Warnings, m.status.Alerts.Errors)
	if contentWidth(m.width) < 32 {
		alerts = fmt.Sprintf("Alertas %d/%d/%d", m.status.Alerts.OK, m.status.Alerts.Warnings, m.status.Alerts.Errors)
	}

	refreshAction := m.statusAction(0, "Atualizar status", "[R]")
	backAction := m.statusAction(1, "Voltar", "[Esc]")
	actions := lipgloss.JoinHorizontal(lipgloss.Top, refreshAction, "  ", backAction)
	if contentWidth(m.width) < 44 {
		refreshLabel := "Atualizar status"
		if contentWidth(m.width) < 28 {
			refreshLabel = "Atualizar"
		}
		actions = lipgloss.JoinVertical(lipgloss.Left, m.statusAction(0, refreshLabel, "[R]"), m.statusAction(1, "Voltar", "[Esc]"))
	}
	meta := fmt.Sprintf("Rede: %s\n%s\n%s", network, installations, alerts)

	view := strings.Join([]string{
		tagStyle.Render("STATUS"),
		titleStyle.Render("Estado do ambiente"),
		statusColor(string(m.status.Overall)).Render("● Ambiente " + overallStateLabel(m.status.Overall)),
		mutedStyle.Render("Verificado agora" + generated),
		joinCards(cards, m.width),
		tagStyle.Render("DETALHES DO AMBIENTE"),
		m.statusTable.View(),
		bodyStyle.Render(meta),
		actions,
	}, "\n\n")
	lines := strings.Split(view, "\n")
	for index := range lines {
		lines[index] = strings.TrimRight(lines[index], " ")
	}
	return strings.Join(lines, "\n")
}

func statusCard(width int, title string, state status.CheckState, detail string) string {
	stateText := checkStateLabel(state)
	content := tagStyle.Render(title) + "\n" + statusColor(string(state)).Render("● "+stateText) + "\n" + mutedStyle.Render(detail)
	cardWidth := contentWidth(width)
	if contentColumns(width) == 2 {
		cardWidth = (cardWidth - 2) / 2
	}
	return cardStyle.Copy().Width(cardWidth).Render(content)
}

func (m Model) statusAction(index int, label, shortcut string) string {
	text := shortcut + " " + label
	if m.focus == index {
		return primaryButtonStyle.Render(text)
	}
	return buttonStyle.Render(text)
}

func checkStateLabel(value status.CheckState) string {
	switch value {
	case status.CheckOK:
		return "ok"
	case status.CheckWarning:
		return "atenção"
	case status.CheckError:
		return "erro"
	case status.CheckMissing:
		return "ausente"
	default:
		return "desconhecido"
	}
}

func overallStateLabel(value status.OverallState) string {
	switch value {
	case status.StateHealthy:
		return "saudável"
	case status.StateDegraded:
		return "atenção"
	case status.StateError:
		return "erro"
	default:
		return "desconhecido"
	}
}

func sshState(value status.SSHStatus) status.CheckState {
	if value.Running {
		return status.CheckOK
	}
	if value.Enabled {
		return status.CheckWarning
	}
	return value.State
}

func sshDetail(value status.SSHStatus) string {
	if value.Running {
		return fmt.Sprintf("porta %d · servidor ativo", value.Port)
	}
	return fmt.Sprintf("porta %d · servidor parado", value.Port)
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func statusNetworkSummary(width int, value status.NetworkStatus) string {
	network := value.Preferred
	if network == "" && len(value.Addresses) > 0 {
		network = value.Addresses[0]
	}
	if network == "" {
		network = valueOr(string(value.State), "indisponível")
	}
	limit := max(8, contentWidth(width)-6)
	runes := []rune(network)
	if len(runes) > limit {
		network = string(runes[:limit-1]) + "…"
	}
	return network
}

func statusRows(value status.SystemStatus, width int) []table.Row {
	columns := statusTableColumns(width)
	stateWidth := max(1, columns[1].Width-2)
	right := func(text string) string {
		return lipgloss.NewStyle().Width(stateWidth).Align(lipgloss.Right).Render(text)
	}
	runtimeLabel := "Termux"
	if !value.Host.Termux {
		runtimeLabel = "Runtime"
	}
	return []table.Row{{runtimeLabel, right(checkStateLabel(value.Host.State))}, {"Arquitetura", right(valueOr(value.Host.Architecture, "desconhecida"))}, {"Ubuntu", right(checkStateLabel(value.Ubuntu.State))}, {"Workspace", right(yesNo(value.Ubuntu.Workspace))}, {"SSH", right(checkStateLabel(value.SSH.State))}, {"Porta SSH", right(fmt.Sprintf("%d", value.SSH.Port))}, {"Wake-lock", right(available(value.Host.WakeLockAvailable))}, {"Bateria", right(batterySummary(value.Battery))}, {"Wi-Fi", right(wifiSummary(value.WiFi))}}
}

func statusTableColumns(width int) []table.Column {
	width = contentWidth(width)
	if width < 32 {
		return []table.Column{{Title: "Item", Width: 8}, {Title: "Estado", Width: 7}}
	}
	itemWidth := min(18, (width-4)/2)
	stateWidth := max(10, width-4-itemWidth)
	stateTitle := lipgloss.NewStyle().Width(max(1, stateWidth-2)).Align(lipgloss.Right).Render("Estado")
	return []table.Column{{Title: "Item", Width: itemWidth}, {Title: stateTitle, Width: stateWidth}}
}

func batterySummary(value status.BatteryStatus) string {
	if value.Percentage == nil {
		return string(value.State)
	}
	return fmt.Sprintf("%d%% %s", *value.Percentage, value.Status)
}
func wifiSummary(value status.WiFiStatus) string {
	if value.IP != "" {
		return value.IP
	}
	if value.Connected {
		return "conectado"
	}
	return string(value.State)
}
