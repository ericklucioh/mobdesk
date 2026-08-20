package tui

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ericklucioh/mobdesk/internal/i18n"
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

func (m Model) appStateLabel(value string) string {
	id := i18n.TUIPopupAppAvailable
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

func popupUsesHelp(entry toolEntry) bool {
	for _, argument := range entry.profile.VersionArg {
		if argument == "--help" || argument == "-h" {
			return true
		}
	}
	return false
}

func popupVersion(entry toolEntry, installation status.InstallationStatus, installed bool) string {
	if !installed {
		return ""
	}
	if popupUsesHelp(entry) {
		return entry.profile.CatalogVersion
	}
	version := strings.TrimSpace(installation.Version)
	if newline := strings.IndexByte(version, '\n'); newline >= 0 {
		version = strings.TrimSpace(version[:newline])
	}
	return version
}

func popupDependencyLabel(name string) string {
	switch name {
	case "go":
		return "Go"
	case "node":
		return "Node.js"
	case "python":
		return "Python"
	default:
		if name == "" {
			return ""
		}
		return strings.ToUpper(name[:1]) + name[1:]
	}
}

func (m Model) popupActions() []popupAction {
	entry, ok := m.popupEntry()
	if !ok {
		return []popupAction{{ID: "close", Label: m.text(i18n.TUIPopupClose, nil), Enabled: true}}
	}
	installation, installed := m.popupInstallation(entry)
	actions := make([]popupAction, 0, 5)
	installLabel := m.text(i18n.TUIPopupInstall, nil)
	if installed && installation.State == "installed" {
		installLabel = m.text(i18n.TUIPopupReinstall, nil)
	}
	if !installed || installation.State != "installed" {
		canInstall := m.canManageHost() && !m.status.Storage.Blocked
		reason := m.hostActionReason(m.canManageHost())
		if m.status.Storage.Blocked {
			reason = m.text(i18n.TUIPopupStorageBlocked, nil)
		}
		actions = append(actions, popupAction{ID: "install", Label: installLabel, Enabled: canInstall, Reason: reason})
		actions = append(actions, popupAction{ID: "close", Label: m.text(i18n.TUIPopupClose, nil), Enabled: true})
		return actions
	}
	canInstall := m.canManageHost() && !m.status.Storage.Blocked
	installReason := m.hostActionReason(m.canManageHost())
	if m.status.Storage.Blocked {
		installReason = m.text(i18n.TUIPopupStorageBlocked, nil)
	}
	actions = append(actions, popupAction{ID: "install", Label: installLabel, Enabled: canInstall, Reason: installReason})
	managed := installed && installation.State == "installed" && installation.Managed && installation.Source == "mobdesk"
	uninstallReason := m.text(i18n.TUIPopupDetectedReason, nil)
	if !installed || installation.State != "installed" {
		uninstallReason = m.text(i18n.TUIPopupInstallFirst, nil)
	} else if managed && m.canManageHost() {
		uninstallReason = ""
	}
	actions = append(actions, popupAction{ID: "uninstall", Label: m.text(i18n.TUIPopupUninstall, nil), Enabled: managed && m.canManageHost(), Destructive: true, Reason: uninstallReason})
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
	state := m.text(i18n.TUIPopupAppAvailable, nil)
	if installed {
		state = m.appStateLabel(installation.State)
	}
	version := popupVersion(entry, installation, installed)
	var builder strings.Builder
	builder.WriteString(titleStyle.Render(toolAppLabel(entry)) + "\n")
	builder.WriteString(popupWrap(entry.profile.Description, width-8) + "\n\n")
	metadata := []string{m.text(i18n.TUIPopupState, map[string]any{"Value": state})}
	if version != "" {
		metadata = append(metadata, m.text(i18n.TUIPopupVersion, map[string]any{"Value": version}))
	}
	if entry.profile.Usage != "" {
		metadata = append(metadata, m.text(i18n.TUIPopupUsage, map[string]any{"Value": entry.profile.Usage}))
	}
	if len(entry.profile.Requires) > 0 {
		dependencies := make([]string, 0, len(entry.profile.Requires))
		for _, dependency := range entry.profile.Requires {
			dependencies = append(dependencies, popupDependencyLabel(dependency))
		}
		metadata = append(metadata, m.text(i18n.TUIPopupDependencies, map[string]any{"Value": strings.Join(dependencies, ", ")}))
	}
	if installed && len(installation.MissingExecutables) > 0 {
		metadata = append(metadata, m.text(i18n.TUIPopupMissingExecutables, map[string]any{"Value": strings.Join(installation.MissingExecutables, ", ")}))
	}
	if installed && len(installation.MissingDependencies) > 0 {
		metadata = append(metadata, m.text(i18n.TUIPopupDependencies, map[string]any{"Value": strings.Join(installation.MissingDependencies, ", ")}))
	}
	if !installed {
		if estimate := entry.profile.StorageEstimate; estimate != nil {
			metadata = append(metadata, m.text(i18n.TUIPopupStorageShort, map[string]any{"Min": estimate.TotalMinMB(), "Max": estimate.TotalMaxMB()}))
		}
	}
	builder.WriteString(popupWrap(strings.Join(metadata, "\n"), width-8) + "\n")
	if m.popupMessage != "" {
		builder.WriteString("\n" + statusColor("warning").Render(popupWrap(m.popupMessage, width-8)) + "\n")
	}
	actions := m.popupActions()
	buttonViews := make([]string, 0, len(actions))
	for index, action := range actions {
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
		buttonViews = append(buttonViews, line)
	}
	if len(buttonViews) > 1 {
		closeButton := buttonViews[len(buttonViews)-1]
		buttonViews = buttonViews[:len(buttonViews)-1]
		canJoin := true
		joinedWidth := 0
		for _, button := range buttonViews {
			if strings.Contains(button, "\n") {
				canJoin = false
				break
			}
			joinedWidth += lipgloss.Width(button)
		}
		if canJoin && joinedWidth+max(0, len(buttonViews)-1)*2 <= width-8 {
			builder.WriteString("\n" + strings.Join(buttonViews, "  ") + "\n")
			buttonViews = nil
		}
		buttonViews = append(buttonViews, closeButton)
	}
	for _, button := range buttonViews {
		builder.WriteString("\n" + button + "\n")
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
		return m.runHostOperation("install", "install", name)
	case "uninstall":
		return m.runHostOperation("uninstall", "uninstall", name, "--json", "--progress")
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
