package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ericklucioh/mobdesk/internal/install"
	"github.com/ericklucioh/mobdesk/internal/status"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case quitAfterMouseResetMsg:
		return m, tea.Quit
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
		// No fallback X10, usado por alguns terminais móveis, o release do
		// botão esquerdo é codificado como MouseNone. O clique inicial ainda
		// precisa ser MouseLeft; aqui basta confirmar que havia um toque ativo.
		if msg.Button != tea.MouseLeft && msg.Button != tea.MouseNone || !m.pointerDown {
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
		if msg.id != 0 && msg.id != m.statusID {
			return m, nil
		}
		if msg.err != nil {
			if !m.busy {
				m.installingTool = ""
				m.message = msg.err.Error()
			}
			return m, nil
		}
		m.status = msg.value
		m.version = msg.info
		m.statusLoaded = true
		m.installingTool = ""
		m.statusTable.SetRows(statusRows(msg.value, m.width))
	case operationMessage:
		if msg.id != 0 && msg.id != m.operationID {
			return m, nil
		}
		m.busy = false
		m.operation = ""
		m.operationProgress = ""
		if msg.command == "update" || msg.command == "update-check" {
			m.systemMessage = operationMessageText(msg)
			m.systemState = operationState(msg)
			m.message = ""
		} else {
			m.message = operationMessageText(msg)
		}
		m.applyOperationState(msg)
		return m.requestStatus()
	case operationProgressMessage:
		if msg.id != 0 && msg.id != m.operationID {
			return m, nil
		}
		m.operationProgress = msg.message
		return m, msg.next
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}
	return m, nil
}

func (m *Model) applyOperationState(msg operationMessage) {
	if msg.err != nil || !msg.result.Success {
		return
	}
	switch msg.command {
	case "start":
		m.status.SSH.Enabled = true
		m.status.SSH.Running = true
		m.status.SSH.State = status.CheckOK
		if msg.result.Port > 0 {
			m.status.SSH.Port = msg.result.Port
		}
		if len(msg.result.Addresses) > 0 {
			m.status.Network.Addresses = append([]string(nil), msg.result.Addresses...)
			m.status.Network.Preferred = msg.result.Addresses[0]
		}
	case "stop":
		m.status.SSH.Running = false
		m.status.SSH.State = status.CheckWarning
	}
}

const hostActionUnavailableMessage = "Esta ação exige o Termux host. Saia da sessão SSH e execute-a no Termux."

func (m Model) canManageHost() bool {
	// Enquanto o status inicial ainda carrega, preserve as ações do host em vez
	// de tratá-las como indisponíveis por causa do valor zero do modelo.
	return !m.statusLoaded || m.status.Host.Termux
}

func (m Model) hostActionUnavailable() (tea.Model, tea.Cmd) {
	m.message = hostActionUnavailableMessage
	return m, nil
}

func (m Model) runHostOperation(operation string, args ...string) (tea.Model, tea.Cmd) {
	if !m.canManageHost() {
		return m.hostActionUnavailable()
	}
	if m.busy {
		return m, nil
	}
	m.operationID++
	m.statusID++ // Invalida snapshots iniciados antes da operação mutável.
	m.busy, m.operation = true, operation
	m.operationProgress = ""
	if operation == "install" && len(args) > 1 {
		m.operationProgress = "Preparando instalação de " + args[1]
	}
	if operation == "update" || operation == "update-check" {
		m.systemMessage, m.systemState = "", ""
	}
	return m, operationCommand(m.operationID, operation, m.backend.OperationCmd(args...))
}

func operationCommand(operationID int, operation string, command tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		message := command()
		switch value := message.(type) {
		case operationMessage:
			value.id = operationID
			return value
		case operationProgressMessage:
			value.id = operationID
			value.next = operationCommand(operationID, operation, value.next)
			return value
		default:
			return operationMessage{command: operation, id: operationID, err: fmt.Errorf("resposta inesperada da operação")}
		}
	}
}

func (m Model) requestStatus() (tea.Model, tea.Cmd) {
	m.statusID++
	return m, m.statusCommand(m.statusID)
}

func (m Model) statusCommand(statusID int) tea.Cmd {
	return func() tea.Msg {
		message, ok := m.backend.StatusCmd()().(statusMessage)
		if !ok {
			return statusMessage{id: statusID, err: fmt.Errorf("resposta inesperada do status")}
		}
		message.id = statusID
		return message
	}
}

func (m Model) cancelBackend() {
	if backend, ok := m.backend.(interface{ Cancel() }); ok {
		backend.Cancel()
	}
}

func (m Model) updateMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	if m.screen == toolsScreen {
		switch msg.Button {
		case tea.MouseWheelDown:
			m.toolsList.CursorDown()
		case tea.MouseWheelUp:
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
		message := msg.result.Message
		if !msg.result.Success && msg.result.LogPath != "" {
			message += "\nLog: " + msg.result.LogPath
		}
		return message
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

func operationState(msg operationMessage) string {
	if msg.err != nil || !msg.result.Success {
		return "failed"
	}
	return msg.result.State
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
		if !m.busy {
			return m.requestStatus()
		}
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
			return m.runHostOperation("update-check", "update", "--check", "--json")
		}
	case "a":
		if m.screen == systemScreen {
			return m.runHostOperation("update", "update", "--json")
		}
	case "e":
		if m.screen == setupScreen {
			return m.runHostOperation("setup-upgrade", "setup", "--upgrade-system", "--json")
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
			return m.runHostOperation("stop", "stop", "--json")
		}
		m.confirmExit = false
		m.closing = true
		m.cancelBackend()
		return m, tea.Tick(time.Millisecond, func(time.Time) tea.Msg {
			return quitAfterMouseResetMsg{}
		})
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
	if !m.canManageHost() {
		return m.hostActionUnavailable()
	}
	if m.busy {
		return m, nil
	}
	entries := toolEntries("")
	index := m.toolsList.Index()
	if index < 0 || index >= len(entries) {
		return m, nil
	}
	m.selectedTool = index
	m.installingTool = entries[index].profile.Name
	return m.runHostOperation("install", "install", entries[index].profile.Name, "--json", "--progress")
}

func (m Model) toggleWorkstation() (tea.Model, tea.Cmd) {
	if !m.canManageHost() {
		return m.hostActionUnavailable()
	}
	if m.busy {
		return m, nil
	}
	if m.status.SSH.Running {
		m.confirmStop = true
		return m, nil
	}
	return m.runHostOperation("start", "start", "--json")
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
		if !m.canManageHost() {
			return 2
		}
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
		if !m.canManageHost() {
			switch m.focus {
			case 0:
				m.navigate(statusScreen)
			case 1:
				m.navigate(shellScreen)
			}
			return nil, true
		}
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
			if m.busy {
				return nil, true
			}
			updated, cmd := m.requestStatus()
			*m = updated.(Model)
			return cmd, true
		}
		m.navigate(homeScreen)
		return nil, true
	case setupScreen:
		if m.focus == 0 {
			updated, cmd := m.runHostOperation("setup", "setup", "--json")
			*m = updated.(Model)
			return cmd, true
		}
		if m.focus == 1 {
			updated, cmd := m.runHostOperation("setup-upgrade", "setup", "--upgrade-system", "--json")
			*m = updated.(Model)
			return cmd, true
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
			updated, cmd := m.runHostOperation("update-check", "update", "--check", "--json")
			*m = updated.(Model)
			return cmd, true
		case 1:
			updated, cmd := m.runHostOperation("update", "update", "--json")
			*m = updated.(Model)
			return cmd, true
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
	if m.closing {
		view.MouseMode = tea.MouseModeNone
	} else {
		view.MouseMode = tea.MouseModeCellMotion
	}
	return view
}

type quitAfterMouseResetMsg struct{}

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
	footer := footerStyle.Width(max(1, terminalWidth(m.width)-2)).Render("↑↓ rolar  Tab foco  Enter agir  R atualizar  Q sair\n" + m.help.View(tuiHelpKeyMap{}))
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
	return headerStyle.Width(headerWidth).Render(line)
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
	style := modalStyle.Width(max(10, contentWidth(width)-6))
	dialog := style.Render(text)
	return lipgloss.PlaceHorizontal(contentWidth(width), lipgloss.Center, dialog)
}
