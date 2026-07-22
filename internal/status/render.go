package status

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

func EncodeJSON(w io.Writer, value SystemStatus) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func RenderText(w io.Writer, value SystemStatus) {
	fmt.Fprintln(w, "Mobdesk status")
	fmt.Fprintf(w, "\nResumo:        %s\n", value.Overall)
	fmt.Fprintf(w, "Atualizado:    %s\n", value.GeneratedAt.Format(time.RFC3339))

	fmt.Fprintln(w, "\nHost")
	fmt.Fprintf(w, "  Termux:      %s\n", value.Host.State)
	fmt.Fprintf(w, "  Arquitetura: %s\n", value.Host.Architecture)
	fmt.Fprintf(w, "  Wake-lock:   %s\n", availability(value.Host.WakeLockAvailable))
	fmt.Fprintf(w, "  Termux:API:  %s\n", availability(value.Host.TermuxAPIAvailable))

	fmt.Fprintln(w, "\nArmazenamento")
	fmt.Fprintf(w, "  Dispositivo: %s livres de %s\n", formatBytes(value.Storage.DeviceFree), formatBytes(value.Storage.DeviceTotal))

	fmt.Fprintln(w, "\nSetup")
	fmt.Fprintf(w, "  Estado:      %s\n", value.Setup.State)
	fmt.Fprintf(w, "  Completo:    %s\n", yesNo(value.Setup.Completed))

	fmt.Fprintln(w, "\nUbuntu")
	fmt.Fprintf(w, "  Estado:      %s\n", value.Ubuntu.State)
	fmt.Fprintf(w, "  Acessível:   %s\n", yesNo(value.Ubuntu.Accessible))
	fmt.Fprintf(w, "  Workspace:   %s\n", yesNo(value.Ubuntu.Workspace))

	fmt.Fprintln(w, "\nSSH")
	fmt.Fprintf(w, "  Estado:      %s\n", value.SSH.State)
	fmt.Fprintf(w, "  Porta:       %d\n", value.SSH.Port)
	fmt.Fprintf(w, "  Rodando:     %s\n", yesNo(value.SSH.Running))

	fmt.Fprintln(w, "\nRede")
	fmt.Fprintf(w, "  Estado:      %s\n", value.Network.State)
	fmt.Fprintf(w, "  Endereços:   %s\n", joinOrUnknown(value.Network.Addresses))

	fmt.Fprintln(w, "\nDispositivo")
	fmt.Fprintf(w, "  Bateria:     %s\n", batteryText(value.Battery))
	fmt.Fprintf(w, "  Wi-Fi:       %s\n", wifiText(value.WiFi))

	if len(value.Installations) > 0 {
		fmt.Fprintln(w, "\nInstalações")
		for _, installation := range value.Installations {
			version := installation.Version
			if version == "" {
				version = installation.State
			}
			fmt.Fprintf(w, "  %s: %s (%s)\n", installation.Name, installation.State, version)
			if installation.LastError != "" {
				fmt.Fprintf(w, "    Erro: %s\n", installation.LastError)
			}
			if installation.LogPath != "" {
				fmt.Fprintf(w, "    Log:  %s\n", installation.LogPath)
			}
		}
	}

	fmt.Fprintln(w, "\nAlertas")
	fmt.Fprintf(w, "  OK: %d | avisos: %d | erros: %d | ausentes: %d | desconhecidos: %d\n",
		value.Alerts.OK, value.Alerts.Warnings, value.Alerts.Errors, value.Alerts.Missing, value.Alerts.Unknown)
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
