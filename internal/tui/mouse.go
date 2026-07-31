package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ericklucioh/mobdesk/internal/i18n"
)

func (m Model) handleMouse(mouse tea.Mouse) (tea.Model, tea.Cmd) {
	if mouse.Button != tea.MouseLeft {
		return m, nil
	}
	if m.confirmExit || m.confirmStop {
		switch m.confirmationMouseAction(mouse) {
		case "yes":
			return m.updateConfirmation("enter")
		case "no":
			m.confirmExit, m.confirmStop = false, false
		}
		return m, nil
	}
	if m.appPopupOpen {
		bodyIndex := m.viewport.YOffset() + mouse.Y - m.bodyTop()
		lines := strings.Split(m.renderScreen(), "\n")
		if m.popupConfirm {
			if action := m.popupConfirmationMouseAction(lines, bodyIndex, mouse.X); action != "" {
				return m.updatePopupKey(action)
			}
			return m, nil
		}
		if focus, ok := m.popupActionAt(lines, bodyIndex, mouse.X); ok {
			m.popupFocus = focus
			return m.activatePopupAction()
		}
		return m, nil
	}
	if mouse.Y < m.bodyTop() {
		headerLines := strings.Split(ansi.Strip(m.renderHeader()), "\n")
		if renderedHeaderLabelAt(headerLines, mouse.Y, mouse.X, "[ "+m.text(i18n.TUIHeaderClose, nil)+" ]") || renderedHeaderLabelAt(headerLines, mouse.Y, mouse.X, "["+m.text(i18n.TUIHeaderClose, nil)+"]") {
			m.confirmExit = true
			return m, nil
		}
		if m.screen != homeScreen && (renderedHeaderLabelAt(headerLines, mouse.Y, mouse.X, "[ "+m.text(i18n.TUIHeaderHome, nil)+" ]") || renderedHeaderLabelAt(headerLines, mouse.Y, mouse.X, "["+firstRune(m.text(i18n.TUIHeaderHome, nil))+"]")) {
			m.navigate(homeScreen)
		}
		return m, nil
	}
	bodyIndex := m.viewport.YOffset() + mouse.Y - m.bodyTop()
	lines := strings.Split(m.renderScreen(), "\n")
	if bodyIndex < 0 || bodyIndex >= len(lines) {
		return m, nil
	}
	switch m.screen {
	case homeScreen:
		if blockContainsAtAny(lines, bodyIndex, mouse.X, m.text(i18n.TUIHomeStart, nil)) || blockContainsAtAny(lines, bodyIndex, mouse.X, m.text(i18n.TUIHomeStop, nil)) {
			return m.toggleWorkstation()
		}
		if blockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUIHomeWorkstationTitle, nil)) {
			return m.toggleWorkstation()
		}
		if blockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUIHomeSetupTitle, nil)) {
			m.navigate(setupScreen)
		} else if blockContainsAtAny(lines, bodyIndex, mouse.X, m.text(i18n.TUIHomeStatusTitle, nil)) {
			m.navigate(statusScreen)
		} else if blockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUIHomeAppsTitle, nil)) {
			m.navigate(toolsScreen)
		} else if blockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUIHomeShellTitle, nil)) {
			m.navigate(shellScreen)
		} else if blockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUIHomeSystemTitle, nil)) {
			m.navigate(systemScreen)
		}
	case toolsScreen:
		for index, entry := range toolEntriesLocalized("", m.localizer) {
			if toolRowContainsAt(lines, bodyIndex, mouse.X, toolAppLabel(entry), contentWidth(m.width)) {
				m.selectedTool = index
				m.toolsList.Select(index)
				m.openAppPopup(index)
				return m, nil
			}
		}
	case setupScreen:
		if touchBlockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUISetupContinue, nil)) {
			return m.runHostOperation("setup", "setup")
		}
		if touchBlockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUISetupUpgrade, nil)) {
			return m.runHostOperation("setup-upgrade", "setup", "--upgrade-system")
		}
		if nearLine(lines, bodyIndex, m.text(i18n.TUISetupUpgrade, nil)) {
			return m.runHostOperation("setup-upgrade", "setup", "--upgrade-system")
		}
	case statusScreen:
		if blockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUIStatusRefresh, nil)) {
			if !m.busy {
				return m.requestStatus()
			}
			return m, nil
		}
		if blockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUIStatusBack, nil)) {
			m.navigate(homeScreen)
		}
	case shellScreen:
		if touchBlockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUIShellOpen, nil)) {
			return m, m.backend.ShellCmd()
		}
		if touchBlockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUIShellBack, nil)) {
			m.navigate(homeScreen)
		}
	case systemScreen:
		if blockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUISystemCheck, nil)) {
			return m.runHostOperation("update-check", "update", "--check", "--json")
		}
		if blockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUISystemUpdate, nil)) {
			return m.runHostOperation("update", "update", "--json")
		}
		if blockContainsAt(lines, bodyIndex, mouse.X, m.text(i18n.TUIStatusBack, nil)) {
			m.navigate(homeScreen)
		}
	}
	return m, nil
}

func (m Model) confirmationMouseAction(mouse tea.Mouse) string {
	lines := strings.Split(ansi.Strip(m.confirmationModal(m.confirmStop, m.width)), "\n")
	for index := range lines {
		for _, candidate := range []struct{ label, action string }{{m.text(i18n.TUIConfirmationYes, nil), "yes"}, {m.text(i18n.TUIConfirmationNo, nil), "no"}} {
			if abs(mouse.Y-(m.bodyTop()+index)) <= 1 && renderedLabelAt(lines, index, mouse.X, candidate.label, 0) {
				return candidate.action
			}
		}
	}
	return ""
}
