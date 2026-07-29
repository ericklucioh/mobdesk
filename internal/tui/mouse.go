package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m Model) handleMouse(mouse tea.Mouse) (tea.Model, tea.Cmd) {
	if mouse.Button != tea.MouseLeft {
		return m, nil
	}
	if m.confirmExit || m.confirmStop {
		switch confirmationMouseAction(mouse, m.width, m.height) {
		case "yes":
			return m.updateConfirmation("enter")
		case "no":
			m.confirmExit, m.confirmStop = false, false
		}
		return m, nil
	}
	if mouse.Y <= 3 {
		if mouse.X >= terminalWidth(m.width)-8 {
			m.confirmExit = true
			return m, nil
		}
		if m.screen != homeScreen && mouse.X >= terminalWidth(m.width)-18 {
			m.navigate(homeScreen)
			return m, nil
		}
		if m.screen != homeScreen && mouse.X >= 12 && mouse.X <= 22 {
			m.navigate(homeScreen)
		}
		return m, nil
	}
	bodyIndex := m.viewport.YOffset() + mouse.Y - 4
	lines := strings.Split(m.renderScreen(), "\n")
	if bodyIndex < 0 || bodyIndex >= len(lines) {
		return m, nil
	}
	switch m.screen {
	case homeScreen:
		if blockContainsAtAny(lines, bodyIndex, mouse.X, "Iniciar") || blockContainsAtAny(lines, bodyIndex, mouse.X, "Parar") {
			return m.toggleWorkstation()
		}
		if blockContainsAt(lines, bodyIndex, mouse.X, "Workstation SSH") {
			return m.toggleWorkstation()
		}
		if blockContainsAt(lines, bodyIndex, mouse.X, "Configurar") {
			m.navigate(setupScreen)
		} else if blockContainsAtAny(lines, bodyIndex, mouse.X, "Status") {
			m.navigate(statusScreen)
		} else if blockContainsAt(lines, bodyIndex, mouse.X, "Apps e linguagens") {
			m.navigate(toolsScreen)
		} else if blockContainsAt(lines, bodyIndex, mouse.X, "Shell Ubuntu") {
			m.navigate(shellScreen)
		} else if blockContainsAt(lines, bodyIndex, mouse.X, "Sistema") {
			m.navigate(systemScreen)
		}
	case toolsScreen:
		for index, entry := range toolEntries("") {
			if toolRowContainsAt(lines, bodyIndex, mouse.X, toolAppLabel(entry), contentWidth(m.width)) {
				m.selectedTool = index
				m.toolsList.Select(index)
				// Toda a linha funciona como alvo no celular. O teclado ainda
				// permite selecionar e confirmar sem depender do mouse.
				return m.installSelectedTool()
			}
		}
	case setupScreen:
		if touchBlockContainsAt(lines, bodyIndex, mouse.X, "Continuar configuração") {
			return m.runHostOperation("setup", "setup", "--json")
		}
		if touchBlockContainsAt(lines, bodyIndex, mouse.X, "upgrade completo") {
			return m.runHostOperation("setup-upgrade", "setup", "--upgrade-system", "--json")
		}
		if nearLine(lines, bodyIndex, "upgrade") || nearLine(lines, bodyIndex, "upgrade completo") {
			return m.runHostOperation("setup-upgrade", "setup", "--upgrade-system", "--json")
		}
		if nearLine(lines, bodyIndex, "Setup retomável") {
			return m.runHostOperation("setup", "setup", "--json")
		}
	case statusScreen:
		if blockContainsAt(lines, bodyIndex, mouse.X, "atualizar") {
			if !m.busy {
				return m.requestStatus()
			}
			return m, nil
		}
		if blockContainsAt(lines, bodyIndex, mouse.X, "voltar") {
			m.navigate(homeScreen)
		}
	case shellScreen:
		if touchBlockContainsAt(lines, bodyIndex, mouse.X, "Abrir shell Ubuntu") {
			return m, m.backend.ShellCmd()
		}
		if touchBlockContainsAt(lines, bodyIndex, mouse.X, "Voltar para início") {
			m.navigate(homeScreen)
		}
	case systemScreen:
		if blockContainsAt(lines, bodyIndex, mouse.X, "Verificar") {
			return m.runHostOperation("update-check", "update", "--check", "--json")
		}
		if blockContainsAt(lines, bodyIndex, mouse.X, "Atualizar") {
			return m.runHostOperation("update", "update", "--json")
		}
		if blockContainsAt(lines, bodyIndex, mouse.X, "Voltar") {
			m.navigate(homeScreen)
		}
	}
	return m, nil
}

func confirmationMouseAction(mouse tea.Mouse, width, height int) string {
	_ = height
	// O modal é renderizado imediatamente abaixo do header (4 linhas). Com
	// borda e padding, a linha dos botões fica na posição 8 do terminal.
	const modalButtonY = 8
	if mouse.Y < modalButtonY-1 || mouse.Y > modalButtonY+1 {
		return ""
	}
	center := width / 2
	if mouse.X >= center-14 && mouse.X <= center-2 {
		return "yes"
	}
	if mouse.X >= center+2 && mouse.X <= center+14 {
		return "no"
	}
	return ""
}
