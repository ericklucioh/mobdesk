package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/status"
)

func (m Model) renderStatus() string {
	if !m.statusLoaded {
		return renderPage(m.text(i18n.TUIStatusTag, nil), m.text(i18n.TUIStatusLoading, nil), m.text(i18n.TUIStatusLoadingDetail, nil))
	}

	generated := ""
	if !m.status.GeneratedAt.IsZero() {
		generated = " · " + m.status.GeneratedAt.Format("15:04")
	}

	hostTitle := m.text(i18n.TUIStatusHost, nil)
	hostDetail := fmt.Sprintf("%s · %s", valueOr(m.status.Host.OS, m.text(i18n.TUIStatusAndroidHost, nil)), valueOr(m.status.Host.Architecture, m.text(i18n.TUIStatusUnknownArchitecture, nil)))
	workspaceDetail := valueOr(m.status.Workspace.Path, m.text(i18n.TUIStatusNetworkUnavailable, nil))
	cards := []string{
		statusCard(m.width, hostTitle, m.status.Host.State, hostDetail),
		statusCard(m.width, m.text(i18n.TUIStatusWorkspace, nil), m.status.Workspace.State, workspaceDetail),
		statusCard(m.width, "SSH", sshState(m.status.SSH), m.sshDetail(m.status.SSH)),
		statusCard(m.width, m.text(i18n.TUIStatusResources, nil), m.status.Storage.State, m.text(i18n.TUIStatusFreeBattery, map[string]any{"Free": formatBytes(m.status.Storage.DeviceFree), "Battery": batterySummaryLocalized(m.status.Battery, m.localizer)})),
	}

	network := statusNetworkSummary(m.width, m.status.Network, m.localizer)
	installations := m.text(i18n.TUIStatusInstallationsCount, map[string]any{"Count": len(m.status.Installations)})
	alerts := m.text(i18n.TUIStatusAlertsCount, map[string]any{"OK": m.status.Alerts.OK, "Warnings": m.status.Alerts.Warnings, "Errors": m.status.Alerts.Errors})
	if contentWidth(m.width) < 32 {
		alerts = m.text(i18n.TUIStatusAlertsShort, map[string]any{"OK": m.status.Alerts.OK, "Warnings": m.status.Alerts.Warnings, "Errors": m.status.Alerts.Errors})
	}

	refreshAction := m.statusAction(0, m.text(i18n.TUIStatusRefresh, nil), "[R]")
	backAction := m.statusAction(1, m.text(i18n.TUIStatusBack, nil), "[Esc]")
	actions := lipgloss.JoinHorizontal(lipgloss.Top, refreshAction, "  ", backAction)
	if contentWidth(m.width) < 44 {
		refreshLabel := m.text(i18n.TUIStatusRefresh, nil)
		if contentWidth(m.width) < 28 {
			refreshLabel = m.text(i18n.TUIStatusRefreshShort, nil)
		}
		actions = lipgloss.JoinVertical(lipgloss.Left, m.statusAction(0, refreshLabel, "[R]"), m.statusAction(1, m.text(i18n.TUIStatusBack, nil), "[Esc]"))
	}
	meta := fmt.Sprintf("%s: %s\n%s\n%s", m.text(i18n.StatusNetwork, nil), network, installations, alerts)

	view := strings.Join([]string{
		tagStyle.Render(m.text(i18n.TUIStatusTag, nil)),
		titleStyle.Render(m.text(i18n.TUIStatusTitle, nil)),
		statusColor(string(m.status.Overall)).Render(ansi.Truncate("● "+m.text(i18n.TUIStatusOverall, map[string]any{"Value": overallStateLabel(m.status.Overall, m.localizer)}), contentWidth(m.width), "…")),
		mutedStyle.Render(m.text(i18n.TUIStatusVerified, map[string]any{"Generated": generated})),
		joinCards(cards, m.width),
		tagStyle.Render(m.text(i18n.TUIStatusDetails, nil)),
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
	cardWidth := max(1, contentWidth(width)-4)
	if contentColumns(width) == 2 {
		cardWidth = max(1, (contentWidth(width)-4)/2)
	}
	return cardStyle.Width(cardWidth).Render(content)
}

func (m Model) statusAction(index int, label, shortcut string) string {
	text := shortcut + " " + label
	if m.focus == index {
		return primaryButtonStyle.Render(text)
	}
	return buttonStyle.Render(text)
}

func checkStateLabel(value status.CheckState, localizers ...i18n.Localizer) string {
	localizer := i18n.New(i18n.LocaleENUS)
	if len(localizers) > 0 {
		localizer = localizers[0]
	}
	switch value {
	case status.CheckOK:
		return localizer.Text(i18n.StatusCheckOK, nil)
	case status.CheckWarning:
		return localizer.Text(i18n.StatusCheckWarning, nil)
	case status.CheckError:
		return localizer.Text(i18n.StatusCheckError, nil)
	case status.CheckMissing:
		return localizer.Text(i18n.StatusCheckMissing, nil)
	default:
		return localizer.Text(i18n.StatusCheckUnknown, nil)
	}
}

func overallStateLabel(value status.OverallState, localizers ...i18n.Localizer) string {
	localizer := i18n.New(i18n.LocaleENUS)
	if len(localizers) > 0 {
		localizer = localizers[0]
	}
	switch value {
	case status.StateHealthy:
		return localizer.Text(i18n.StatusOverallHealthy, nil)
	case status.StateDegraded:
		return localizer.Text(i18n.StatusOverallDegraded, nil)
	case status.StateError:
		return localizer.Text(i18n.StatusOverallError, nil)
	default:
		return localizer.Text(i18n.StatusOverallUnknown, nil)
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

func (m Model) sshDetail(value status.SSHStatus) string {
	if value.Running {
		return m.text(i18n.TUIStatusSSHRunning, map[string]any{"Port": value.Port})
	}
	return m.text(i18n.TUIStatusSSHStopped, map[string]any{"Port": value.Port})
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func statusNetworkSummary(width int, value status.NetworkStatus, localizers ...i18n.Localizer) string {
	localizer := i18n.New(i18n.LocaleENUS)
	if len(localizers) > 0 {
		localizer = localizers[0]
	}
	network := value.Preferred
	if network == "" && len(value.Addresses) > 0 {
		network = value.Addresses[0]
	}
	if network == "" {
		network = valueOr("", localizer.Text(i18n.TUIStatusNetworkUnavailable, nil))
	}
	limit := max(8, contentWidth(width)-6)
	runes := []rune(network)
	if len(runes) > limit {
		network = string(runes[:limit-1]) + "…"
	}
	return network
}

func statusRows(value status.SystemStatus, width int, localizers ...i18n.Localizer) []table.Row {
	localizer := i18n.New(i18n.LocaleENUS)
	if len(localizers) > 0 {
		localizer = localizers[0]
	}
	columns := statusTableColumns(width, localizer)
	stateWidth := max(1, columns[1].Width-2)
	right := func(text string) string {
		return lipgloss.NewStyle().Width(stateWidth).Align(lipgloss.Right).Render(text)
	}
	rows := []table.Row{{localizer.Text(i18n.TUIStatusHost, nil), right(checkStateLabel(value.Host.State, localizer))}, {localizer.Text(i18n.TUIStatusArchitecture, nil), right(valueOr(value.Host.Architecture, localizer.Text(i18n.TUIStatusUnknownArchitecture, nil)))}, {localizer.Text(i18n.TUIStatusWorkspace, nil), right(checkStateLabel(value.Workspace.State, localizer))}, {localizer.Text(i18n.StatusSSH, nil), right(checkStateLabel(value.SSH.State, localizer))}, {localizer.Text(i18n.TUIStatusSSHPort, nil), right(fmt.Sprintf("%d", value.SSH.Port))}, {localizer.Text(i18n.TUIStatusWakeLock, nil), right(availableLocalized(value.Host.WakeLockAvailable, localizer))}, {localizer.Text(i18n.TUIStatusBattery, nil), right(batterySummaryLocalized(value.Battery, localizer))}, {localizer.Text(i18n.TUIStatusWiFi, nil), right(wifiSummaryLocalized(value.WiFi, localizer))}}
	if value.Java.Installed {
		rows = append(rows, table.Row{localizer.Text(i18n.StatusJava, nil), right(valueOr(value.Java.Version, checkStateLabel(value.Java.State, localizer)))})
	}
	return rows
}

func statusTableColumns(width int, localizers ...i18n.Localizer) []table.Column {
	localizer := i18n.New(i18n.LocaleENUS)
	if len(localizers) > 0 {
		localizer = localizers[0]
	}
	width = contentWidth(width)
	if width < 32 {
		return []table.Column{{Title: localizer.Text(i18n.TUIStatusItem, nil), Width: 8}, {Title: localizer.Text(i18n.TUIStatusTableState, nil), Width: 7}}
	}
	itemWidth := min(18, (width-4)/2)
	stateWidth := max(10, width-4-itemWidth)
	stateTitle := lipgloss.NewStyle().Width(max(1, stateWidth-2)).Align(lipgloss.Right).Render(localizer.Text(i18n.TUIStatusTableState, nil))
	return []table.Column{{Title: localizer.Text(i18n.TUIStatusItem, nil), Width: itemWidth}, {Title: stateTitle, Width: stateWidth}}
}

func batterySummary(value status.BatteryStatus) string {
	return batterySummaryLocalized(value, i18n.New(i18n.LocaleENUS))
}
func batterySummaryLocalized(value status.BatteryStatus, localizer i18n.Localizer) string {
	if value.Percentage == nil {
		return checkStateLabel(value.State, localizer)
	}
	statusText := value.Status
	if value.Status == "normal" {
		statusText = localizer.Text(i18n.TUIStatusBatteryNormal, nil)
	} else if value.Status == "low" {
		statusText = localizer.Text(i18n.TUIStatusBatteryLow, nil)
	}
	return fmt.Sprintf("%d%% %s", *value.Percentage, statusText)
}
func wifiSummary(value status.WiFiStatus) string {
	return wifiSummaryLocalized(value, i18n.New(i18n.LocaleENUS))
}
func wifiSummaryLocalized(value status.WiFiStatus, localizer i18n.Localizer) string {
	if value.IP != "" {
		return value.IP
	}
	if value.Connected {
		return localizer.Text(i18n.TUIStatusConnected, nil)
	}
	return string(value.State)
}
