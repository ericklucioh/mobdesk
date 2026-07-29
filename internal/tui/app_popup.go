package tui

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/ericklucioh/mobdesk/internal/status"
)

type popupAction struct {
	ID          string
	Label       string
	Enabled     bool
	Destructive bool
	Reason      string
}

func (m Model) popupEntry() (toolEntry, bool) {
	entries := toolEntriesLocalized("", m.localizer)
	if m.popupAppIndex < 0 || m.popupAppIndex >= len(entries) {
		return toolEntry{}, false
	}
	return entries[m.popupAppIndex], true
}

func (m Model) popupInstallation(entry toolEntry) (status.InstallationStatus, bool) {
	for _, installation := range m.status.Installations {
		if installation.Name == entry.profile.Name || installation.Package == entry.profile.Package || installation.Executable == entry.profile.Executable {
			return installation, true
		}
	}
	return status.InstallationStatus{}, false
}

func (m Model) popupConfiguration(entry toolEntry) (status.ConfigurationStatus, bool) {
	for _, configuration := range m.status.Configurations {
		if configuration.App == entry.profile.Name {
			return configuration, true
		}
	}
	return status.ConfigurationStatus{}, false
}

func (m Model) configStateLabel(value status.ConfigState) string {
	id := i18n.TUIPopupConfigState
	switch value {
	case status.ConfigStateUnavailable:
		id = i18n.TUIPopupConfigUnavailableState
	case status.ConfigStateNotApplied:
		id = i18n.TUIPopupConfigNotApplied
	case status.ConfigStateApplying:
		id = i18n.TUIPopupConfigApplying
	case status.ConfigStateApplied:
		id = i18n.TUIPopupConfigApplied
	case status.ConfigStateRemoving:
		id = i18n.TUIPopupConfigRemoving
	case status.ConfigStateRemoved:
		id = i18n.TUIPopupConfigRemoved
	case status.ConfigStateModified:
		id = i18n.TUIPopupConfigModified
	case status.ConfigStateConflict:
		id = i18n.TUIPopupConfigConflict
	case status.ConfigStateFailed:
		id = i18n.TUIPopupConfigFailed
	}
	return m.text(id, nil)
}

func (m Model) appStateLabel(value string) string {
	id := i18n.TUIPopupConfigState
	switch value {
	case "available":
		id = i18n.TUIPopupAppStateAvailable
	case "installing":
		id = i18n.TUIPopupAppStateInstalling
	case "installed":
		id = i18n.TUIPopupAppStateInstalled
	case "uninstalling":
		id = i18n.TUIPopupAppStateUninstalling
	case "uninstalled":
		id = i18n.TUIPopupAppStateUninstalled
	case "partial":
		id = i18n.TUIPopupAppStatePartial
	case "failed":
		id = i18n.TUIPopupAppStateFailed
	}
	return m.text(id, map[string]any{"Value": value})
}

func (m Model) popupActions() []popupAction {
	entry, ok := m.popupEntry()
	if !ok {
		return []popupAction{{ID: "close", Label: m.text(i18n.TUIPopupClose, nil), Enabled: true}}
	}
	installation, installed := m.popupInstallation(entry)
	configuration, hasConfig := m.popupConfiguration(entry)
	if !hasConfig && entry.profile.ConfigProfile != "" {
		configuration = status.ConfigurationStatus{App: entry.profile.Name, Profile: entry.profile.ConfigProfile, State: status.ConfigStateNotApplied}
		hasConfig = true
	}
	actions := make([]popupAction, 0, 5)
	if !installed || installation.State != "installed" {
		actions = append(actions, popupAction{ID: "install", Label: m.text(i18n.TUIPopupInstall, nil), Enabled: m.canManageHost(), Reason: m.hostActionReason(m.canManageHost())})
	} else {
		reason := m.text(i18n.TUIPopupAlreadyInstalled, nil)
		if !m.canManageHost() {
			reason = m.text(i18n.TUIHostRestriction, nil)
		}
		actions = append(actions, popupAction{ID: "install", Label: m.text(i18n.TUIPopupInstall, nil), Reason: reason})
	}
	managed := installed && installation.State == "installed" && installation.Managed && installation.Source == "mobdesk"
	uninstallReason := m.text(i18n.TUIPopupDetectedReason, nil)
	if !installed || installation.State != "installed" {
		uninstallReason = m.text(i18n.TUIPopupInstallFirst, nil)
	} else if managed && m.canManageHost() {
		uninstallReason = ""
	}
	actions = append(actions, popupAction{ID: "uninstall", Label: m.text(i18n.TUIPopupUninstall, nil), Enabled: managed && m.canManageHost(), Destructive: true, Reason: uninstallReason})
	if entry.profile.ConfigProfile != "" {
		reason := ""
		enabled := installed && installation.State == "installed" && hasConfig && configuration.State != status.ConfigStateApplied && configuration.State != status.ConfigStateModified && configuration.State != status.ConfigStateConflict && m.canManageHost()
		if !installed || installation.State != "installed" {
			reason = m.text(i18n.TUIPopupInstallFirst, nil)
		} else if configuration.State == status.ConfigStateConflict {
			reason = m.text(i18n.TUIPopupConflict, nil)
		} else if !m.canManageHost() {
			reason = m.text(i18n.TUIHostRestriction, nil)
		}
		actions = append(actions, popupAction{ID: "config_apply", Label: m.text(i18n.TUIPopupApplyConfig, nil), Enabled: enabled, Reason: reason})
		removeEnabled := installed && (configuration.State == status.ConfigStateApplied || configuration.State == status.ConfigStateModified) && m.canManageHost()
		removeReason := m.text(i18n.TUIPopupNotApplied, nil)
		if !m.canManageHost() {
			removeReason = m.text(i18n.TUIHostRestriction, nil)
		}
		actions = append(actions, popupAction{ID: "config_remove", Label: m.text(i18n.TUIPopupRemoveConfig, nil), Enabled: removeEnabled, Destructive: true, Reason: removeReason})
	}
	actions = append(actions, popupAction{ID: "close", Label: m.text(i18n.TUIPopupClose, nil), Enabled: true})
	return actions
}

func (m Model) hostActionReason(manageable bool) string {
	if manageable {
		return ""
	}
	return m.text(i18n.TUIHostRestriction, nil)
}

func popupWrap(text string, width int) string {
	return lipgloss.NewStyle().Width(max(1, width)).Render(text)
}

func popupActionLabel(action popupAction, width int) string {
	return popupActionLabelLocalized(action, width, i18n.New(i18n.LocaleENUS))
}

func popupActionLabelLocalized(action popupAction, width int, localizer i18n.Localizer) string {
	if width < 32 {
		switch action.ID {
		case "uninstall":
			return localizer.Text(i18n.TUIPopupUninstallShort, nil)
		case "config_apply":
			return localizer.Text(i18n.TUIPopupApplyConfigShort, nil)
		case "config_remove":
			return localizer.Text(i18n.TUIPopupRemoveConfigShort, nil)
		}
	}
	return "[ " + action.Label + " ]"
}

func (m Model) renderAppPopup() string {
	entry, ok := m.popupEntry()
	if !ok {
		return ""
	}
	width := contentWidth(m.width)
	installation, installed := m.popupInstallation(entry)
	configuration, hasConfig := m.popupConfiguration(entry)
	if !hasConfig && entry.profile.ConfigProfile != "" {
		configuration = status.ConfigurationStatus{State: status.ConfigStateNotApplied, Profile: entry.profile.ConfigProfile}
		hasConfig = true
	}
	state, source, version := m.text(i18n.TUIPopupAppAvailable, nil), m.text(i18n.TUIPopupNotInstalled, nil), m.text(i18n.TUIPopupNotDetected, nil)
	if installed {
		state, source = m.appStateLabel(installation.State), installation.Source
		if source == "mobdesk" {
			source = m.text(i18n.TUIPopupMobdesk, nil)
		} else if source == "detected" {
			source = m.text(i18n.TUIPopupDetected, nil)
		}
		if installation.Version != "" {
			version = installation.Version
		}
	}
	var builder strings.Builder
	builder.WriteString(tagStyle.Render(m.text(i18n.TUIPopupTag, nil)) + "\n")
	builder.WriteString(titleStyle.Render(toolAppLabel(entry)) + "\n")
	builder.WriteString(popupWrap(entry.profile.Description, width-8) + "\n\n")
	builder.WriteString(popupWrap(strings.Join([]string{m.text(i18n.TUIPopupState, map[string]any{"Value": state}), m.text(i18n.TUIPopupSource, map[string]any{"Value": source}), m.text(i18n.TUIPopupVersion, map[string]any{"Value": version})}, "\n"), width-8) + "\n")
	if len(entry.profile.Requires) > 0 {
		builder.WriteString(popupWrap(m.text(i18n.TUIPopupDependencies, map[string]any{"Value": strings.Join(entry.profile.Requires, ", ")}), width-8) + "\n")
	}
	if hasConfig {
		builder.WriteString(popupWrap(m.text(i18n.TUIPopupConfig, map[string]any{"Value": m.configStateLabel(configuration.State)}), width-8) + "\n")
		if len(configuration.ManagedPaths) > 0 {
			builder.WriteString(popupWrap(m.text(i18n.TUIPopupPaths, map[string]any{"Value": strings.Join(configuration.ManagedPaths, ", ")}), width-8) + "\n")
		}
		profile := install.DefaultConfigProfiles()[entry.profile.ConfigProfile]
		if len(profile.ManagedPlugins) > 0 {
			builder.WriteString(popupWrap(m.text(i18n.TUIPopupPlugins, map[string]any{"Count": len(profile.ManagedPlugins)}), width-8) + "\n")
		}
	} else {
		builder.WriteString(popupWrap(m.text(i18n.TUIPopupConfigUnavailable, nil), width-8) + "\n")
	}
	if estimate := entry.profile.StorageEstimate; estimate != nil {
		builder.WriteString(popupWrap(m.text(i18n.TUIPopupStorage, map[string]any{"AppMin": estimate.AppMinMB, "AppMax": estimate.AppMaxMB, "DepMin": estimate.DependenciesMinMB, "DepMax": estimate.DependenciesMaxMB, "ConfigMin": estimate.ConfigMinMB, "ConfigMax": estimate.ConfigMaxMB}), width-8) + "\n")
		builder.WriteString(popupWrap(m.text(i18n.TUIPopupStorageTotal, map[string]any{"Min": estimate.TotalMinMB(), "Max": estimate.TotalMaxMB()}), width-8) + "\n")
	}
	if m.popupMessage != "" {
		builder.WriteString("\n" + statusColor("warning").Render(popupWrap(m.popupMessage, width-8)) + "\n")
	}
	builder.WriteString("\n" + m.text(i18n.TUIPopupActions, nil) + "\n")
	for index, action := range m.popupActions() {
		label := popupActionLabelLocalized(action, width, m.localizer)
		style := mutedStyle
		if action.Enabled {
			style = buttonStyle
			if index == m.popupFocus {
				style = primaryButtonStyle
			}
		}
		line := style.Render(label)
		if !action.Enabled && action.Reason != "" {
			line += "\n" + mutedStyle.Render(popupWrap(action.Reason, width-8))
		}
		builder.WriteString(line + "\n")
	}
	if m.popupConfirm {
		builder.WriteString("\n" + modalStyle.Render(m.text(i18n.TUIPopupConfirm, nil)+"\n\n"+m.text(i18n.TUIConfirmationYes, nil)+"     "+m.text(i18n.TUIConfirmationNo, nil)))
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, modalStyle.Width(max(10, width-8)).Render(builder.String()))
}

func (m *Model) openAppPopup(index int) {
	entries := toolEntriesLocalized("", m.localizer)
	if index < 0 || index >= len(entries) {
		return
	}
	m.popupAppIndex = index
	m.popupFocus = 0
	m.popupAction = ""
	m.popupConfirm = false
	m.popupMessage = ""
	m.appPopupOpen = true
}

func (m *Model) closeAppPopup() {
	m.appPopupOpen = false
	m.popupConfirm = false
	m.popupAction = ""
	m.popupMessage = ""
}

func (m Model) updatePopupKey(key string) (tea.Model, tea.Cmd) {
	if m.popupConfirm {
		switch key {
		case "y", "Y", "s", "S", "enter":
			action := m.popupAction
			m.popupConfirm = false
			m.popupAction = ""
			return m.dispatchPopupAction(action)
		case "n", "N", "esc":
			m.popupConfirm = false
			m.popupAction = ""
			m.popupMessage = m.text(i18n.TUIActionCancelled, nil)
		}
		return m, nil
	}
	actions := m.popupActions()
	switch key {
	case "tab":
		m.popupFocus = (m.popupFocus + 1) % len(actions)
	case "shift+tab":
		m.popupFocus = (m.popupFocus + len(actions) - 1) % len(actions)
	case "enter":
		return m.activatePopupAction()
	case "esc":
		m.closeAppPopup()
	}
	return m, nil
}

func (m Model) activatePopupAction() (tea.Model, tea.Cmd) {
	actions := m.popupActions()
	if m.popupFocus < 0 || m.popupFocus >= len(actions) {
		return m, nil
	}
	action := actions[m.popupFocus]
	if !action.Enabled {
		m.popupMessage = action.Reason
		return m, nil
	}
	if action.ID == "close" {
		m.closeAppPopup()
		return m, nil
	}
	if action.Destructive {
		m.popupAction = action.ID
		m.popupConfirm = true
		m.popupMessage = ""
		return m, nil
	}
	return m.dispatchPopupAction(action.ID)
}

func (m Model) dispatchPopupAction(action string) (tea.Model, tea.Cmd) {
	entry, ok := m.popupEntry()
	if !ok || !m.canManageHost() {
		return m.hostActionUnavailable()
	}
	name := entry.profile.Name
	switch action {
	case "install":
		m.installingTool = name
		return m.runHostOperation("install", "install", name, "--json", "--progress")
	case "uninstall":
		return m.runHostOperation("uninstall", "uninstall", name, "--json", "--progress")
	case "config_apply":
		return m.runHostOperation("config-apply", "config", "apply", name, "--json", "--progress")
	case "config_remove":
		return m.runHostOperation("config-remove", "config", "remove", name, "--json", "--progress")
	default:
		return m, nil
	}
}

func (m Model) popupActionAt(lines []string, bodyIndex, x int) (int, bool) {
	actions := m.popupActions()
	for index, action := range actions {
		label := popupActionLabelLocalized(action, contentWidth(m.width), m.localizer)
		for position, line := range lines {
			plain := ansi.Strip(line)
			start := strings.Index(plain, label)
			if start < 0 || bodyIndex < position-1 || bodyIndex > position+1 {
				continue
			}
			first := utf8.RuneCountInString(plain[:start])
			last := first + utf8.RuneCountInString(label)
			if x >= first && x <= last {
				return index, true
			}
		}
	}
	return 0, false
}

func (m Model) popupConfirmationMouseAction(lines []string, bodyIndex, x int) string {
	for _, candidate := range []struct {
		label  string
		action string
	}{
		{label: m.text(i18n.TUIConfirmationYes, nil), action: "y"},
		{label: m.text(i18n.TUIConfirmationNo, nil), action: "n"},
	} {
		for position, line := range lines {
			plain := ansi.Strip(line)
			start := strings.Index(plain, candidate.label)
			if start < 0 || bodyIndex < position-1 || bodyIndex > position+1 {
				continue
			}
			first := utf8.RuneCountInString(plain[:start])
			last := first + utf8.RuneCountInString(candidate.label)
			if x >= first && x <= last {
				return candidate.action
			}
		}
	}
	return ""
}
