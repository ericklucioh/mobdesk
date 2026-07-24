package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	"github.com/ericklucioh/mobdesk/internal/status"
)

func (m Model) renderStatus() string {
	if !m.statusLoaded {
		return renderPage("STATUS", "Carregando", "Coletando estado do ambiente...")
	}
	intro := fmt.Sprintf("Resumo: %s\n%s\n%s\n\nRede: %s\nArmazenamento: %s livres\nInstalações: %d\nAlertas: %d OK · %d avisos · %d erros", m.status.Overall, m.focusAction(0, "[R] atualizar status"), m.focusAction(1, "[Esc] voltar"), strings.Join(m.status.Network.Addresses, ", "), formatBytes(m.status.Storage.DeviceFree), len(m.status.Installations), m.status.Alerts.OK, m.status.Alerts.Warnings, m.status.Alerts.Errors)
	return tagStyle.Render("STATUS") + "\n" + titleStyle.Render("Estado do ambiente") + "\n" + bodyStyle.Render(intro) + "\n\n" + m.statusTable.View()
}

func statusRows(value status.SystemStatus) []table.Row {
	return []table.Row{{"Termux", string(value.Host.State)}, {"Arquitetura", value.Host.Architecture}, {"Ubuntu", string(value.Ubuntu.State)}, {"Workspace", yesNo(value.Ubuntu.Workspace)}, {"SSH", string(value.SSH.State)}, {"Porta SSH", fmt.Sprintf("%d", value.SSH.Port)}, {"Wake-lock", available(value.Host.WakeLockAvailable)}, {"Bateria", batterySummary(value.Battery)}, {"Wi-Fi", wifiSummary(value.WiFi)}}
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
