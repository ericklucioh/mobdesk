package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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
	entries := toolEntries("")
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

func (m Model) popupActions() []popupAction {
	entry, ok := m.popupEntry()
	if !ok {
		return []popupAction{{ID: "close", Label: "Fechar", Enabled: true}}
	}
	installation, installed := m.popupInstallation(entry)
	configuration, hasConfig := m.popupConfiguration(entry)
	if !hasConfig && entry.profile.ConfigProfile != "" {
		configuration = status.ConfigurationStatus{App: entry.profile.Name, Profile: entry.profile.ConfigProfile, State: status.ConfigStateNotApplied}
		hasConfig = true
	}
	actions := make([]popupAction, 0, 5)
	if !installed || installation.State != "installed" {
		actions = append(actions, popupAction{ID: "install", Label: "Instalar", Enabled: m.canManageHost(), Reason: hostActionReason(m.canManageHost())})
	} else {
		reason := "O app já está instalado"
		if !m.canManageHost() {
			reason = hostActionUnavailableMessage
		}
		actions = append(actions, popupAction{ID: "install", Label: "Instalar", Reason: reason})
	}
	managed := installed && installation.State == "installed" && installation.Managed && installation.Source == "mobdesk"
	uninstallReason := "O app foi apenas detectado; não há proveniência segura"
	if !installed || installation.State != "installed" {
		uninstallReason = "Instale o app antes de desinstalá-lo"
	} else if managed && m.canManageHost() {
		uninstallReason = ""
	}
	actions = append(actions, popupAction{ID: "uninstall", Label: "Desinstalar", Enabled: managed && m.canManageHost(), Destructive: true, Reason: uninstallReason})
	if entry.profile.ConfigProfile != "" {
		reason := ""
		enabled := installed && installation.State == "installed" && hasConfig && configuration.State != status.ConfigStateApplied && configuration.State != status.ConfigStateModified && configuration.State != status.ConfigStateConflict && m.canManageHost()
		if !installed || installation.State != "installed" {
			reason = "Instale o app antes de aplicar a configuração"
		} else if configuration.State == status.ConfigStateConflict {
			reason = "A configuração existente gera conflito"
		} else if !m.canManageHost() {
			reason = hostActionUnavailableMessage
		}
		actions = append(actions, popupAction{ID: "config_apply", Label: "Adicionar configuração Mobdesk", Enabled: enabled, Reason: reason})
		removeEnabled := installed && (configuration.State == status.ConfigStateApplied || configuration.State == status.ConfigStateModified) && m.canManageHost()
		removeReason := "A configuração ainda não foi aplicada"
		if !m.canManageHost() {
			removeReason = hostActionUnavailableMessage
		}
		actions = append(actions, popupAction{ID: "config_remove", Label: "Remover configuração Mobdesk", Enabled: removeEnabled, Destructive: true, Reason: removeReason})
	}
	actions = append(actions, popupAction{ID: "close", Label: "Fechar", Enabled: true})
	return actions
}

func hostActionReason(manageable bool) string {
	if manageable {
		return ""
	}
	return hostActionUnavailableMessage
}

func popupWrap(text string, width int) string {
	return lipgloss.NewStyle().Width(max(1, width)).Render(text)
}

func popupActionLabel(action popupAction, width int) string {
	if width < 32 {
		switch action.ID {
		case "uninstall":
			return "[ Remover app ]"
		case "config_apply":
			return "[ Aplicar config ]"
		case "config_remove":
			return "[ Remover config ]"
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
	state, source, version := "disponível", "não instalado", "não detectada"
	if installed {
		state, source = installation.State, installation.Source
		if source == "mobdesk" {
			source = "Mobdesk"
		} else if source == "detected" {
			source = "detectado"
		}
		if installation.Version != "" {
			version = installation.Version
		}
	}
	var builder strings.Builder
	builder.WriteString(tagStyle.Render("DETALHES DO APP") + "\n")
	builder.WriteString(titleStyle.Render(toolAppLabel(entry)) + "\n")
	builder.WriteString(popupWrap(entry.profile.Description, width-8) + "\n\n")
	builder.WriteString(popupWrap(fmt.Sprintf("Estado: %s\nOrigem: %s\nVersão: %s", state, source, version), width-8) + "\n")
	if len(entry.profile.Requires) > 0 {
		builder.WriteString(popupWrap("Dependências: "+strings.Join(entry.profile.Requires, ", "), width-8) + "\n")
	}
	if hasConfig {
		builder.WriteString(popupWrap(fmt.Sprintf("Configuração Mobdesk: %s", configuration.State), width-8) + "\n")
		if len(configuration.ManagedPaths) > 0 {
			builder.WriteString(popupWrap("Caminhos: "+strings.Join(configuration.ManagedPaths, ", "), width-8) + "\n")
		}
		profile := install.DefaultConfigProfiles()[entry.profile.ConfigProfile]
		if len(profile.ManagedPlugins) > 0 {
			builder.WriteString(popupWrap(fmt.Sprintf("Plugins gerenciados: %d", len(profile.ManagedPlugins)), width-8) + "\n")
		}
	} else {
		builder.WriteString(popupWrap("Configuração Mobdesk: indisponível", width-8) + "\n")
	}
	if estimate := entry.profile.StorageEstimate; estimate != nil {
		builder.WriteString(popupWrap(fmt.Sprintf("Armazenamento: app %d-%d MB · dependências %d-%d MB · config %d-%d MB", estimate.AppMinMB, estimate.AppMaxMB, estimate.DependenciesMinMB, estimate.DependenciesMaxMB, estimate.ConfigMinMB, estimate.ConfigMaxMB), width-8) + "\n")
		builder.WriteString(popupWrap(fmt.Sprintf("Total estimado: %d-%d MB", estimate.TotalMinMB(), estimate.TotalMaxMB()), width-8) + "\n")
	}
	if m.popupMessage != "" {
		builder.WriteString("\n" + statusColor("warning").Render(popupWrap(m.popupMessage, width-8)) + "\n")
	}
	builder.WriteString("\nAções\n")
	for index, action := range m.popupActions() {
		label := popupActionLabel(action, width)
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
		builder.WriteString("\n" + modalStyle.Render("Confirmar ação destrutiva?\n\n[ Y ] Sim     [ N ] Não"))
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, modalStyle.Width(max(10, width-8)).Render(builder.String()))
}

func (m *Model) openAppPopup(index int) {
	entries := toolEntries("")
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
			m.popupMessage = "Ação cancelada"
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
		label := popupActionLabel(action, contentWidth(m.width))
		for position, line := range lines {
			plain := ansi.Strip(line)
			start := strings.Index(plain, label)
			if start < 0 || bodyIndex != position {
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

func popupConfirmationMouseAction(lines []string, bodyIndex, x int) string {
	for _, candidate := range []struct {
		label  string
		action string
	}{
		{label: "[ Y ] Sim", action: "y"},
		{label: "[ N ] Não", action: "n"},
	} {
		for position, line := range lines {
			plain := ansi.Strip(line)
			start := strings.Index(plain, candidate.label)
			if start < 0 || position != bodyIndex {
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
