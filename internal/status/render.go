package status

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ericklucioh/mobdesk/internal/i18n"
)

func EncodeJSON(w io.Writer, value SystemStatus) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func RenderText(w io.Writer, value SystemStatus, localizers ...i18n.Localizer) error {
	localizer := i18n.New(i18n.LocaleENUS)
	if len(localizers) > 0 {
		localizer = localizers[0]
	}
	var text strings.Builder
	text.WriteString(localizer.Text(i18n.StatusTitle, nil) + "\n")
	text.WriteString("\n" + localizer.Text(i18n.StatusSummary, map[string]any{"Value": localizedOverall(localizer, value.Overall)}) + "\n")
	text.WriteString(localizer.Text(i18n.StatusUpdated, map[string]any{"Value": value.GeneratedAt.Format(time.RFC3339)}) + "\n")

	hostLabel := localizer.Text(i18n.StatusRuntime, nil)
	if value.Host.Termux {
		hostLabel = "Termux"
	}
	text.WriteString("\n" + localizer.Text(i18n.StatusHost, nil) + "\n")
	fmt.Fprintf(&text, "  %-12s%s\n", hostLabel+":", localizedCheck(localizer, value.Host.State))
	fmt.Fprintf(&text, "  %s: %s\n  %s:   %s\n  %s:  %s\n", localizer.Text(i18n.StatusArchitecture, nil), value.Host.Architecture, localizer.Text(i18n.StatusWakeLock, nil), availability(localizer, value.Host.WakeLockAvailable), localizer.Text(i18n.StatusTermuxAPI, nil), availability(localizer, value.Host.TermuxAPIAvailable))
	fmt.Fprintf(&text, "\n%s\n  %s\n", localizer.Text(i18n.StatusStorage, nil), localizer.Text(i18n.StatusDeviceStorage, map[string]any{"Free": formatBytes(value.Storage.DeviceFree), "Total": formatBytes(value.Storage.DeviceTotal)}))
	fmt.Fprintf(&text, "\n%s\n  %s:      %s\n  %s:    %s\n", localizer.Text(i18n.StatusSetup, nil), localizer.Text(i18n.StatusState, nil), localizedCheck(localizer, value.Setup.State), localizer.Text(i18n.StatusComplete, nil), yesNo(localizer, value.Setup.Completed))
	fmt.Fprintf(&text, "\n%s\n  %s:      %s\n  %s:       %s\n", localizer.Text(i18n.StatusWorkspace, nil), localizer.Text(i18n.StatusState, nil), localizedCheck(localizer, value.Workspace.State), localizer.Text(i18n.StatusPath, nil), value.Workspace.Path)
	fmt.Fprintf(&text, "\n%s\n  %s:      %s\n  %s:       %d\n  %s:    %s\n", localizer.Text(i18n.StatusSSH, nil), localizer.Text(i18n.StatusState, nil), localizedCheck(localizer, value.SSH.State), localizer.Text(i18n.StatusPort, nil), value.SSH.Port, localizer.Text(i18n.StatusRunning, nil), yesNo(localizer, value.SSH.Running))
	fmt.Fprintf(&text, "\n%s\n  %s:      %s\n  %s:   %s\n", localizer.Text(i18n.StatusNetwork, nil), localizer.Text(i18n.StatusState, nil), localizedCheck(localizer, value.Network.State), localizer.Text(i18n.StatusAddresses, nil), joinOrUnknown(localizer, value.Network.Addresses))
	fmt.Fprintf(&text, "\n%s\n  %s:     %s\n  %s:       %s\n", localizer.Text(i18n.StatusDevice, nil), localizer.Text(i18n.StatusBattery, nil), batteryText(localizer, value.Battery), localizer.Text(i18n.StatusWiFi, nil), wifiText(localizer, value.WiFi))

	if len(value.Installations) > 0 {
		text.WriteString("\n" + localizer.Text(i18n.StatusInstallations, nil) + "\n")
		for _, installation := range value.Installations {
			version := installation.Version
			if version == "" {
				version = installation.State
			}
			fmt.Fprintf(&text, "  %s: %s (%s)\n", installation.Name, localizeState(localizer, installation.State), version)
			if installation.LastError != "" {
				fmt.Fprintf(&text, "    %s: %s\n", localizer.Text(i18n.StatusError, nil), installation.LastError)
			}
			if installation.LogPath != "" {
				fmt.Fprintf(&text, "    %s:  %s\n", localizer.Text(i18n.StatusLog, nil), installation.LogPath)
			}
		}
	}
	fmt.Fprintf(&text, "\n%s\n  %s\n", localizer.Text(i18n.StatusAlerts, nil), localizer.Text(i18n.StatusAlertCounts, map[string]any{"OK": value.Alerts.OK, "Warnings": value.Alerts.Warnings, "Errors": value.Alerts.Errors, "Missing": value.Alerts.Missing, "Unknown": value.Alerts.Unknown}))
	_, err := io.WriteString(w, text.String())
	return err
}

func localizedOverall(l i18n.Localizer, state OverallState) string {
	ids := map[OverallState]i18n.MessageID{StateHealthy: i18n.StatusOverallHealthy, StateDegraded: i18n.StatusOverallDegraded, StateError: i18n.StatusOverallError, StateUnknown: i18n.StatusOverallUnknown}
	if id, ok := ids[state]; ok {
		return l.Text(id, nil)
	}
	return string(state)
}

func localizedCheck(l i18n.Localizer, state CheckState) string {
	ids := map[CheckState]i18n.MessageID{CheckOK: i18n.StatusCheckOK, CheckWarning: i18n.StatusCheckWarning, CheckError: i18n.StatusCheckError, CheckMissing: i18n.StatusCheckMissing, CheckUnknown: i18n.StatusCheckUnknown}
	if id, ok := ids[state]; ok {
		return l.Text(id, nil)
	}
	return string(state)
}

func localizeState(l i18n.Localizer, state string) string {
	return l.Text(i18n.StatusAppState, map[string]any{"Value": state})
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

func availability(l i18n.Localizer, value bool) string {
	if value {
		return l.Text(i18n.StatusAvailable, nil)
	}
	return l.Text(i18n.StatusMissing, nil)
}

func yesNo(l i18n.Localizer, value bool) string {
	if value {
		return l.Text(i18n.StatusYes, nil)
	}
	return l.Text(i18n.StatusNo, nil)
}

func joinOrUnknown(l i18n.Localizer, values []string) string {
	if len(values) == 0 {
		return l.Text(i18n.StatusNone, nil)
	}
	return strings.Join(values, ", ")
}

func batteryText(l i18n.Localizer, value BatteryStatus) string {
	if value.State == CheckMissing {
		return l.Text(i18n.StatusBatteryAPIMissing, nil)
	}
	if value.Percentage == nil {
		return localizedCheck(l, value.State)
	}
	return fmt.Sprintf("%d%% (%s)", *value.Percentage, value.Status)
}

func wifiText(l i18n.Localizer, value WiFiStatus) string {
	if value.State == CheckMissing {
		return l.Text(i18n.StatusBatteryAPIMissing, nil)
	}
	if !value.Connected {
		return l.Text(i18n.StatusDisconnected, nil)
	}
	if value.IP != "" {
		return fmt.Sprintf("%s (%s)", l.Text(i18n.StatusConnected, nil), value.IP)
	}
	return l.Text(i18n.StatusConnected, nil)
}
