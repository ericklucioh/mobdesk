package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ericklucioh/mobdesk/internal/install"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
	case tea.MouseWheelMsg:
		return m.updateMouseWheel(msg)
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			m.pointerDown = true
			m.dragging = false
			m.pressMouse = msg.Mouse()
			m.dragY = msg.Y
		}
		return m, nil
	case tea.MouseMotionMsg:
		if !m.pointerDown {
			return m, nil
		}
		mouse := msg.Mouse()
		if mouse.Button == tea.MouseNone {
			m.pointerDown = false
			m.dragging = false
			return m, nil
		}
		if !m.dragging && (abs(mouse.Y-m.pressMouse.Y) >= 2 || abs(mouse.X-m.pressMouse.X) >= 2) {
			m.dragging = true
		}
		return m.updateMouseDrag(mouse)
	case tea.MouseReleaseMsg:
		if msg.Button != tea.MouseLeft || !m.pointerDown {
			return m, nil
		}
		click := m.pressMouse
		wasDragging := m.dragging
		m.pointerDown = false
		m.dragging = false
		if wasDragging {
			return m, nil
		}
		return m.handleMouse(click)
	case tea.BlurMsg:
		m.pointerDown = false
		m.dragging = false
		return m, nil
	case statusMessage:
		if msg.err != nil {
			m.busy = false
			m.installingTool = ""
			m.message = msg.err.Error()
			return m, nil
		}
		m.status = msg.value
		m.version = msg.info
		m.statusLoaded = true
		m.busy = false
		m.installingTool = ""
		m.statusTable.SetRows(statusRows(msg.value, m.width))
	case operationMessage:
		m.busy = false
		m.operation = ""
		m.message = operationMessageText(msg)
		return m, m.backend.StatusCmd()
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}
	return m, nil
}

func (m Model) updateMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	if m.screen == toolsScreen {
		if msg.Button == tea.MouseWheelDown {
			m.toolsList.CursorDown()
		} else if msg.Button == tea.MouseWheelUp {
			m.toolsList.CursorUp()
		}
		m.selectedTool = m.toolsList.Index()
		return m, nil
	}
	m.viewport.SetContent(m.renderScreen())
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) resize(width, height int) {
	m.width, m.height = width, height
	height = max(6, height-8)
	m.toolsList.SetSize(contentWidth(width), height)
	m.setupActions.SetSize(max(1, contentWidth(width)-4), max(3, min(5, height)))
	m.statusTable.SetColumns(statusTableColumns(width))
	m.statusTable.SetWidth(contentWidth(width))
	// As linhas seguintes pertencem ao resumo da tela, não à tabela. Limitar a
	// altura evita que o componente reserve um painel vazio até o rodapé.
	m.statusTable.SetHeight(min(height, 10))
	m.viewport.SetWidth(contentWidth(width))
	m.viewport.SetHeight(max(1, m.height-3))
}

func operationMessageText(msg operationMessage) string {
	if msg.err != nil {
		return msg.err.Error()
	}
	if msg.result.Message != "" {
		return msg.result.Message
	}
	switch msg.command {
	case "update":
		return "Atualização concluída"
	case "update-check":
		return "Verificação concluída"
	case "install":
		message := msg.result.Language + " instalado"
		if msg.result.Version != "" {
			message += " (" + msg.result.Version + ")"
		}
		return message
	default:
		return ""
	}
}

func (m Model) updateKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if m.screen == toolsScreen && isListKey(key) {
		m.moveSelector(&m.toolsList, key)
		m.selectedTool = m.toolsList.Index()
		return m, nil
	}
	if m.screen == setupScreen && isListKey(key) {
		m.moveSelector(&m.setupActions, key)
		m.focus = m.setupActions.Index()
		return m, nil
	}
	if m.screen != toolsScreen && isViewportKey(key) {
		m.viewport.SetContent(m.renderScreen())
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	if m.confirmExit || m.confirmStop {
		return m.updateConfirmation(key)
	}
	switch key {
	case "tab", "shift+tab":
		count := m.controlCount()
		if count > 0 {
			if key == "shift+tab" {
				m.focus = (m.focus + count - 1) % count
			} else {
				m.focus = (m.focus + 1) % count
			}
		}
	case "esc":
		if len(m.history) > 0 {
			m.screen = m.history[len(m.history)-1]
			m.history = m.history[:len(m.history)-1]
			m.focus = 0
		}
	case "q", "ctrl+c", "x":
		m.confirmExit = true
	case "r":
		return m, m.backend.StatusCmd()
	case "1", "h":
		m.navigate(homeScreen)
	case "2", "d":
		m.navigate(statusScreen)
	case "3", "c":
		m.navigate(setupScreen)
	case "4", "t":
		m.navigate(toolsScreen)
	case "5":
		m.navigate(shellScreen)
	case "6", "u":
		m.navigate(systemScreen)
	case "v":
		if m.screen == systemScreen {
			m.busy, m.operation = true, "update-check"
			return m, m.backend.OperationCmd("update", "--check", "--json")
		}
	case "a":
		if m.screen == systemScreen {
			m.busy, m.operation = true, "update"
			return m, m.backend.OperationCmd("update", "--json")
		}
	case "e":
		if m.screen == setupScreen {
			m.busy, m.operation = true, "setup-upgrade"
			return m, m.backend.OperationCmd("setup", "--upgrade-system", "--json")
		}
	case "i":
		if m.screen == toolsScreen {
			return m.installSelectedTool()
		}
	case "s":
		if m.screen == homeScreen {
			return m.toggleWorkstation()
		}
	case "enter":
		if cmd, handled := m.activateFocusedControl(); handled {
			return m, cmd
		}
	}
	return m, nil
}

func (m *Model) moveSelector(selector *selector, key string) {
	switch key {
	case "up", "k":
		selector.CursorUp()
	case "down", "j":
		selector.CursorDown()
	case "pgup":
		for i := 0; i < max(1, selector.height/4); i++ {
			selector.CursorUp()
		}
	case "pgdown":
		for i := 0; i < max(1, selector.height/4); i++ {
			selector.CursorDown()
		}
	}
}

func (m Model) updateConfirmation(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "Y", "s", "S", "enter":
		if m.confirmStop {
			m.confirmStop = false
			m.busy, m.operation = true, "stop"
			return m, m.backend.OperationCmd("stop", "--json")
		}
		m.confirmExit = false
		return m, tea.Quit
	case "n", "N", "esc", "q":
		m.confirmExit, m.confirmStop = false, false
	}
	return m, nil
}

func isListKey(value string) bool {
	return value == "up" || value == "down" || value == "k" || value == "j" || value == "pgup" || value == "pgdown"
}
func isViewportKey(value string) bool { return isListKey(value) || value == "home" || value == "end" }

func (m Model) installSelectedTool() (tea.Model, tea.Cmd) {
	entries := toolEntries("language")
	index := m.toolsList.Index()
	if index < 0 || index >= len(entries) {
		return m, nil
	}
	m.selectedTool = index
	m.busy, m.operation = true, "install"
	m.installingTool = entries[index].language.Name
	return m, m.backend.OperationCmd("install", entries[index].language.Name, "--json")
}

func (m Model) toggleWorkstation() (tea.Model, tea.Cmd) {
	if m.status.SSH.Running {
		m.confirmStop = true
		return m, nil
	}
	m.busy, m.operation = true, "start"
	return m, m.backend.OperationCmd("start", "--json")
}

func (m *Model) navigate(next screen) {
	if m.screen != next {
		m.history = append(m.history, m.screen)
	}
	m.screen, m.focus = next, 0
	m.viewport.SetYOffset(0)
}

func (m Model) controlCount() int {
	switch m.screen {
	case homeScreen:
		return 7
	case statusScreen, shellScreen:
		return 2
	case systemScreen:
		return 3
	case setupScreen:
		return 2
	case toolsScreen:
		return len(install.Languages())
	default:
		return 0
	}
}

func (m *Model) activateFocusedControl() (tea.Cmd, bool) {
	switch m.screen {
	case homeScreen:
		switch m.focus {
		case 0:
			_, cmd := m.toggleWorkstation()
			return cmd, true
		case 1:
			m.navigate(setupScreen)
		case 2:
			m.navigate(statusScreen)
		case 3:
			m.navigate(toolsScreen)
		case 4:
			m.navigate(shellScreen)
		case 5:
			m.navigate(systemScreen)
		}
		return nil, true
	case statusScreen:
		if m.focus == 0 {
			return m.backend.StatusCmd(), true
		}
		m.navigate(homeScreen)
		return nil, true
	case setupScreen:
		if m.focus == 0 {
			m.busy, m.operation = true, "setup"
			return m.backend.OperationCmd("setup", "--json"), true
		}
		if m.focus == 1 {
			m.busy, m.operation = true, "setup-upgrade"
			return m.backend.OperationCmd("setup", "--upgrade-system", "--json"), true
		}
		return nil, true
	case toolsScreen:
		_, cmd := m.installSelectedTool()
		return cmd, true
	case shellScreen:
		if m.focus == 0 {
			return m.backend.ShellCmd(), true
		}
		m.navigate(homeScreen)
		return nil, true
	case systemScreen:
		switch m.focus {
		case 0:
			m.busy, m.operation = true, "update-check"
			return m.backend.OperationCmd("update", "--check", "--json"), true
		case 1:
			m.busy, m.operation = true, "update"
			return m.backend.OperationCmd("update", "--json"), true
		case 2:
			m.navigate(homeScreen)
		}
		return nil, true
	}
	return nil, false
}

func (m Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	// CellMotion habilita clique, release, roda e movimento durante o arraste.
	// É o modo mais compatível com terminais móveis como o Termux: não envia
	// eventos de movimento contínuos quando o dedo está apenas parado na tela.
	view.MouseMode = tea.MouseModeCellMotion
	return view
}

func (m Model) render() string {
	header := m.renderHeader()
	bodyModel := m
	bodyWidth := contentWidth(m.width)
	bodyModel.viewport.Style = lipgloss.NewStyle().Align(lipgloss.Left)
	bodyModel.viewport.SetWidth(bodyWidth)
	bodyModel.viewport.SetHeight(max(1, m.height-3))
	// A TUI mobile não deve manter deslocamento horizontal entre telas. Linhas
	// longas são tratadas pela própria tela, enquanto o viewport fica ancorado
	// na borda esquerda.
	bodyModel.viewport.SetXOffset(0)
	bodyModel.viewport.SetContent(bodyModel.renderScreen())
	body := bodyModel.viewport.View()
	if m.message != "" && !m.busy && m.screen != homeScreen && !m.confirmExit && !m.confirmStop {
		body += "\n\n" + statusColor("warning").Render(m.message)
	}
	if m.confirmExit || m.confirmStop {
		body = confirmationModal(m.confirmStop, m.width)
	}
	footer := footerStyle.Copy().Width(max(1, terminalWidth(m.width)-2)).Render("↑↓ rolar  Tab foco  Enter agir  R atualizar  Q sair\n" + m.help.View(tuiHelpKeyMap{}))
	return header + "\n" + body + "\n" + footer
}

func (m Model) renderHeader() string {
	headerWidth := max(1, terminalWidth(m.width)-2)
	brand := headerBrandStyle.Render("M/ MOBDESK")
	stateText := "● workstation parada"
	stateStyle := statusColor("stopped")
	if m.status.SSH.Running {
		stateText = "● workstation ativa"
		stateStyle = statusColor("running")
	}
	state := stateStyle.Render(stateText)
	close := headerCloseStyle.Render("[ X ]")
	home := ""
	if m.screen != homeScreen {
		home = headerLinkStyle.Render("[ HOME ]")
	}
	right := home
	if right != "" {
		right += " "
	}
	right += close

	lineWidth := max(1, headerWidth-2) // headerStyle has horizontal padding.
	available := lineWidth - lipgloss.Width(brand) - lipgloss.Width(state) - lipgloss.Width(right)
	if available < 2 {
		// O celular vertical prioriza os quatro elementos do app-bar e encurta
		// somente o texto de estado, nunca quebrando a linha.
		state = stateStyle.Render("● ON")
		if !m.status.SSH.Running {
			state = stateStyle.Render("● OFF")
		}
		if m.screen != homeScreen {
			home = headerLinkStyle.Render("[H]")
		}
		close = headerCloseStyle.Render("[X]")
		right = home
		if right != "" {
			right += " "
		}
		right += close
		available = lineWidth - lipgloss.Width(brand) - lipgloss.Width(state) - lipgloss.Width(right)
	}
	if available < 2 {
		brand = headerBrandStyle.Render("M/")
		available = lineWidth - lipgloss.Width(brand) - lipgloss.Width(state) - lipgloss.Width(right)
	}
	if available < 0 {
		available = 0
	}
	leftGap := available / 2
	rightGap := available - leftGap
	line := brand + strings.Repeat(" ", leftGap) + state + strings.Repeat(" ", rightGap) + right
	return headerStyle.Copy().Width(headerWidth).Render(line)
}

func (m Model) renderScreen() string {
	if m.busy {
		return m.renderOperation()
	}
	switch m.screen {
	case statusScreen:
		return m.renderStatus()
	case setupScreen:
		return m.renderSetup()
	case toolsScreen:
		return m.renderTools()
	case shellScreen:
		return m.renderShell()
	case systemScreen:
		return m.renderSystem()
	default:
		return m.renderHome()
	}
}

func confirmationModal(stop bool, width int) string {
	text := "Fechar TUI?\n\n[ S ] Sim     [ N ] Não"
	if stop {
		text = "Parar workstation?\n\n[ S ] Sim     [ N ] Não"
	}
	style := modalStyle.Copy().Width(max(10, contentWidth(width)-6))
	dialog := style.Render(text)
	return lipgloss.PlaceHorizontal(contentWidth(width), lipgloss.Center, dialog)
}

func screenLabel(value screen) string {
	switch value {
	case homeScreen:
		return "Início"
	case statusScreen:
		return "Status"
	case setupScreen:
		return "Configurar"
	case toolsScreen:
		return "Apps"
	case shellScreen:
		return "Shell"
	case systemScreen:
		return "Sistema"
	default:
		return "Mobdesk"
	}
}
