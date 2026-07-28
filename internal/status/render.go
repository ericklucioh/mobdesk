package status

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

func EncodeJSON(w io.Writer, value SystemStatus) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func RenderText(w io.Writer, value SystemStatus) error {
	var text strings.Builder
	text.WriteString("Mobdesk status\n")
	summary := fmt.Sprintf("\nResumo:        %s\n", value.Overall)
	text.WriteString(summary)
	updated := fmt.Sprintf("Atualizado:    %s\n", value.GeneratedAt.Format(time.RFC3339))
	text.WriteString(updated)

	hostLabel := "Termux"
	if !value.Host.Termux {
		hostLabel = "Runtime"
	}
	text.WriteString("\nHost\n")
	hostState := fmt.Sprintf("  %-12s%s\n", hostLabel+":", value.Host.State)
	text.WriteString(hostState)
	appendStatusf(&text, "  Arquitetura: %s\n  Wake-lock:   %s\n  Termux:API:  %s\n", value.Host.Architecture, availability(value.Host.WakeLockAvailable), availability(value.Host.TermuxAPIAvailable))
	appendStatusf(&text, "\nArmazenamento\n  Dispositivo: %s livres de %s\n", formatBytes(value.Storage.DeviceFree), formatBytes(value.Storage.DeviceTotal))
	appendStatusf(&text, "\nSetup\n  Estado:      %s\n  Completo:    %s\n", value.Setup.State, yesNo(value.Setup.Completed))
	appendStatusf(&text, "\nUbuntu\n  Estado:      %s\n  Acessível:   %s\n  Workspace:   %s\n", value.Ubuntu.State, yesNo(value.Ubuntu.Accessible), yesNo(value.Ubuntu.Workspace))
	appendStatusf(&text, "\nSSH\n  Estado:      %s\n  Porta:       %d\n  Rodando:     %s\n", value.SSH.State, value.SSH.Port, yesNo(value.SSH.Running))
	appendStatusf(&text, "\nRede\n  Estado:      %s\n  Endereços:   %s\n", value.Network.State, joinOrUnknown(value.Network.Addresses))
	appendStatusf(&text, "\nDispositivo\n  Bateria:     %s\n  Wi-Fi:       %s\n", batteryText(value.Battery), wifiText(value.WiFi))

	if len(value.Installations) > 0 {
		text.WriteString("\nInstalações\n")
		for _, installation := range value.Installations {
			version := installation.Version
			if version == "" {
				version = installation.State
			}
			appendStatusf(&text, "  %s: %s (%s)\n", installation.Name, installation.State, version)
			if installation.LastError != "" {
				appendStatusf(&text, "    Erro: %s\n", installation.LastError)
			}
			if installation.LogPath != "" {
				appendStatusf(&text, "    Log:  %s\n", installation.LogPath)
			}
		}
	}

	appendStatusf(&text, "\nAlertas\n  OK: %d | avisos: %d | erros: %d | ausentes: %d | desconhecidos: %d\n",
		value.Alerts.OK, value.Alerts.Warnings, value.Alerts.Errors, value.Alerts.Missing, value.Alerts.Unknown)
	_, err := io.WriteString(w, text.String())
	return err
}

func appendStatusf(builder *strings.Builder, format string, values ...any) {
	formatted := fmt.Sprintf(format, values...)
	builder.WriteString(formatted)
}

func formatBytes(value int64) string {
	if value < 1024 {
		return fmt.Sprintf("%d B", value)
	}
	units := []string{"KB", "MB", "GB", "TB"}
	number := float64(value)
	for _, unit := range units {
		number /= 1024
		if number < 1024 || unit == units[len(units)-1] {
			return fmt.Sprintf("%.1f %s", number, unit)
		}
	}
	return fmt.Sprintf("%d B", value)
}

func availability(value bool) string {
	if value {
		return "disponível"
	}
	return "ausente"
}

func yesNo(value bool) string {
	if value {
		return "sim"
	}
	return "não"
}

func joinOrUnknown(values []string) string {
	if len(values) == 0 {
		return "nenhum"
	}
	result := values[0]
	for _, value := range values[1:] {
		result += ", " + value
	}
	return result
}

func batteryText(value BatteryStatus) string {
	if value.State == CheckMissing {
		return "Termux:API ausente"
	}
	if value.Percentage == nil {
		return string(value.State)
	}
	return fmt.Sprintf("%d%% (%s)", *value.Percentage, value.Status)
}

func wifiText(value WiFiStatus) string {
	if value.State == CheckMissing {
		return "Termux:API ausente"
	}
	if !value.Connected {
		return "desconectado"
	}
	if value.IP != "" {
		return fmt.Sprintf("conectado (%s)", value.IP)
	}
	return "conectado"
}
