package tui

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ericklucioh/mobdesk/internal/status"
)

func TestHomeRendersWorkstationCard(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 50
	model.height = 30
	view := model.renderScreen()
	for _, expected := range []string{"INÍCIO", "Workstation SSH", "Status:", "desativado", "Apps e linguagens"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("home view does not contain %q: %s", expected, view)
		}
	}
}

func TestToolsUseCurrentCatalog(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 50
	view := ansi.Strip(model.renderToolsBubbles())
	for _, expected := range []string{"go", "python3", "nodejs", "clang", "lua5.4"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("tools view does not contain catalog item %q: %s", expected, view)
		}
	}
}

func TestToolsFitVerticalPhoneWidth(t *testing.T) {
	model := New()
	model.width = 20
	model.height = 30
	for _, line := range strings.Split(model.renderTools(), "\n") {
		if lipgloss.Width(line) > contentWidth(model.width) {
			t.Fatalf("tools line exceeds narrow terminal: %q", line)
		}
	}
}

func TestToolsUseBorderedTwoLineItemsAndInstallationState(t *testing.T) {
	model := New()
	model.width = 80
	model.height = 40
	view := ansi.Strip(model.renderTools())
	if !strings.Contains(view, "┌") || !strings.Contains(view, "instalar") {
		t.Fatalf("tools view does not render bordered install items: %s", view)
	}
	lines := strings.Split(view, "\n")
	goLine := -1
	for index, line := range lines {
		if strings.Contains(line, "go") {
			goLine = index
			if !strings.Contains(line, "instalar") {
				t.Fatalf("tool state is not on the App line: %q", line)
			}
			break
		}
	}
	if goLine < 0 || goLine+1 >= len(lines) || !strings.Contains(lines[goLine+1], "Linguagem compilada") {
		t.Fatalf("tool phrase is not directly below App: %q", lines)
	}
	model.statusLoaded = true
	model.status.Host.Termux = true
	model.status.Installations = []status.InstallationStatus{{Name: "go", Kind: "language", State: "installed"}}
	if !toolInstalled(model.status, toolEntries("language")[0]) {
		t.Fatal("installed go was not recognized")
	}
	view = ansi.Strip(model.renderTools())
	if !strings.Contains(view, "instalado") {
		t.Fatalf("tools view does not render installed state: %s", view)
	}
}

func TestToolsShowInstallingUntilFinalStatus(t *testing.T) {
	model := New()
	model.screen = toolsScreen
	model.width = 44
	model.height = 30
	model.resize(model.width, model.height)

	started, _ := model.installSelectedTool()
	model = started.(Model)
	if model.installingTool != "go" || !model.busy {
		t.Fatalf("install did not enter transient state: installing=%q busy=%v", model.installingTool, model.busy)
	}
	if view := ansi.Strip(model.renderTools()); !strings.Contains(view, "Instalando") {
		t.Fatalf("tools view does not show transient installation state: %s", view)
	}

	updated, _ := model.Update(operationMessage{command: "install", result: operationResult{Language: "go"}})
	model = updated.(Model)
	if model.busy || model.installingTool != "go" {
		t.Fatalf("operation result cleared state before status snapshot: busy=%v installing=%q", model.busy, model.installingTool)
	}

	value := status.SystemStatus{
		Host:          status.HostStatus{Termux: true},
		Installations: []status.InstallationStatus{{Name: "go", Kind: "language", State: "installed"}},
	}
	updated, _ = model.Update(statusMessage{value: value})
	model = updated.(Model)
	if model.installingTool != "" || !strings.Contains(ansi.Strip(model.renderTools()), "instalado") {
		t.Fatalf("final status did not settle installation state: installing=%q view=%s", model.installingTool, model.renderTools())
	}
}

func TestOperationViewHasNoFakeProgress(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 30
	model.busy = true
	model.operation = "start"

	view := ansi.Strip(model.renderOperation())
	for _, unexpected := range []string{"Verificando setup", "●", "○", "%", "━"} {
		if strings.Contains(view, unexpected) {
			t.Fatalf("operation view still contains fake progress %q: %s", unexpected, view)
		}
	}
	for _, expected := range []string{"Iniciando workstation", "Operação em andamento", "Aguarde a conclusão"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("operation view does not contain %q: %s", expected, view)
		}
	}
}

func TestInstallOperationViewShowsReportedProgress(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 30
	model.busy = true
	model.operation = "install"
	model.operationID = 3
	next := func() tea.Msg { return operationMessage{command: "install", result: operationResult{Success: true}} }

	updated, command := model.Update(operationProgressMessage{id: 3, message: "Instalando node", next: next})
	model = updated.(Model)
	if command == nil || model.operationProgress != "Instalando node" {
		t.Fatalf("installation progress was not retained: command=%v progress=%q", command != nil, model.operationProgress)
	}
	if !strings.Contains(ansi.Strip(model.renderOperation()), "Instalando node") {
		t.Fatalf("operation view does not show installation progress: %s", model.renderOperation())
	}
}

func TestInstallFailureIncludesLogPath(t *testing.T) {
	message := operationMessageText(operationMessage{command: "install", result: operationResult{
		Success: false,
		Message: "apt-get falhou",
		LogPath: "/private/install.log",
	}})
	if message != "apt-get falhou\nLog: /private/install.log" {
		t.Fatalf("installation failure message = %q", message)
	}
}

func TestToolsMouseWheelMovesBubbleList(t *testing.T) {
	model := New()
	model.screen = toolsScreen
	model.width = 44
	model.height = 12
	model.resize(model.width, model.height)
	for range toolEntries("") {
		updated, _ := model.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
		model = updated.(Model)
	}
	if model.toolsList.Index() != len(toolEntries(""))-1 {
		t.Fatalf("mouse wheel stopped at %d, want last catalog item", model.toolsList.Index())
	}
}

func TestToolsCanSelectAndInstallLastCatalogItem(t *testing.T) {
	model := New()
	model.screen = toolsScreen
	model.width = 44
	model.height = 30
	entries := toolEntries("")
	last := len(entries) - 1

	model.toolsList.Select(last)
	if model.toolsList.Index() != last {
		t.Fatalf("tools selector stopped at %d, want last catalog item %d", model.toolsList.Index(), last)
	}
	if !strings.Contains(ansi.Strip(model.renderTools()), toolAppLabel(entries[last])) {
		t.Fatalf("tools view does not render last catalog item %q", entries[last].language.Name)
	}

	updated, cmd := model.installSelectedTool()
	model = updated.(Model)
	if cmd == nil || model.installingTool != entries[last].language.Name {
		t.Fatalf("last catalog item was not installable: cmd=%v installing=%q", cmd != nil, model.installingTool)
	}
}

func TestReleaseToolRowsDispatchInstallThroughButton(t *testing.T) {
	for _, name := range []string{"lazygit", "leetgo"} {
		t.Run(name, func(t *testing.T) {
			backend := &controlledBackend{}
			model := NewWithBackend(backend)
			model.screen = toolsScreen
			model.statusLoaded = true
			model.status.Host.Termux = true
			model.width = 44
			model.height = 30
			model.resize(model.width, model.height)

			index := -1
			for candidate, entry := range toolEntries("") {
				if entry.language.Name == name {
					index = candidate
					break
				}
			}
			if index < 0 {
				t.Fatalf("%s missing from tool entries", name)
			}
			model.toolsList.Select(index)
			lines := strings.Split(ansi.Strip(model.renderScreen()), "\n")
			row := -1
			for candidate, line := range lines {
				if strings.Contains(line, name) {
					row = candidate
					break
				}
			}
			if row < 0 {
				t.Fatalf("%s row was not rendered", name)
			}

			updated, _ := model.Update(tea.MouseClickMsg{X: 1, Y: row + 4, Button: tea.MouseLeft})
			model = updated.(Model)
			updated, command := model.Update(tea.MouseReleaseMsg{X: 1, Y: row + 4, Button: tea.MouseLeft})
			model = updated.(Model)
			if command == nil || model.installingTool != name {
				t.Fatalf("button did not start %s: command=%v installing=%q", name, command != nil, model.installingTool)
			}
			command()
			if !slices.Equal(backend.operationArgs, []string{"install", name, "--json", "--progress"}) {
				t.Fatalf("button sent args %v for %s", backend.operationArgs, name)
			}
		})
	}
}

func TestViewEnablesTermuxMouseTracking(t *testing.T) {
	view := New().View()
	if !view.AltScreen {
		t.Fatal("TUI must use the alternate screen")
	}
	if view.MouseMode != tea.MouseModeCellMotion {
		t.Fatalf("TUI mouse mode = %v, want CellMotion", view.MouseMode)
	}
}

func TestSetupRendersResponsiveSections(t *testing.T) {
	model := New()
	model.width = 40
	model.height = 40
	view := model.renderSetup()
	for _, expected := range []string{"PRIMEIRO ACESSO", "Configurar Mobdesk", "Diretórios do Mobdesk", "Workspace e SSH", "OPÇÃO AVANÇADA"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("setup view does not contain %q: %s", expected, view)
		}
	}
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > contentWidth(model.width) {
			t.Fatalf("setup line exceeds terminal width: %q", line)
		}
	}
}

func TestShellRendersLargeTouchActions(t *testing.T) {
	model := New()
	model.width = 40
	model.height = 30
	view := model.renderShell()
	for _, expected := range []string{"Abrir shell Ubuntu", "Suspender a TUI", "Voltar para início", "Retornar à tela principal"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("shell view does not contain %q: %s", expected, view)
		}
	}
	if strings.Count(view, "┌") < 2 || strings.Count(view, "└") < 2 {
		t.Fatalf("shell actions are not rendered as bordered touch targets: %s", view)
	}
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > contentWidth(model.width) {
			t.Fatalf("shell line exceeds terminal width: %q", line)
		}
	}
}

func TestStatusRendersResponsiveSections(t *testing.T) {
	model := New()
	model.statusLoaded = true
	model.width = 20
	model.height = 30
	model.status = status.SystemStatus{
		Overall: status.StateHealthy,
		Host:    status.HostStatus{State: status.CheckOK, Termux: true, OS: "Android", Architecture: "arm64"},
		Ubuntu:  status.UbuntuStatus{State: status.CheckOK, Workspace: true},
		SSH:     status.SSHStatus{State: status.CheckWarning, Port: 8022},
		Storage: status.StorageStatus{State: status.CheckOK, DeviceFree: 42 * 1024 * 1024 * 1024},
		Battery: status.BatteryStatus{State: status.CheckOK, Percentage: intPtr(72), Status: "normal"},
		Network: status.NetworkStatus{Addresses: []string{"172.19.0.1", "172.18.0.1", "10.42.0.1"}},
	}
	model.resize(model.width, model.height)
	view := model.renderStatus()
	for _, expected := range []string{"STATUS", "Estado do ambiente", "TERMUX", "DETALHES DO AMBIENTE", "Atualizar"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("status view does not contain %q: %s", expected, view)
		}
	}
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > contentWidth(model.width) {
			t.Fatalf("status line exceeds narrow terminal: %q", line)
		}
	}
}

func TestRemoteRuntimeHidesHostActions(t *testing.T) {
	model := New()
	model.statusLoaded = true
	model.status.Host.Termux = false
	model.width = 44
	model.height = 30

	home := ansi.Strip(model.renderHome())
	for _, expected := range []string{"Sessão Ubuntu remota", "Status", "Shell Ubuntu"} {
		if !strings.Contains(home, expected) {
			t.Fatalf("remote home does not contain %q: %s", expected, home)
		}
	}
	if strings.Contains(home, "Iniciar") || strings.Contains(home, "Configurar") {
		t.Fatalf("remote home still exposes host actions: %s", home)
	}

	for _, screenView := range []string{model.renderSetup(), model.renderTools(), model.renderSystem()} {
		if !strings.Contains(ansi.Strip(screenView), "Termux") {
			t.Fatalf("remote screen does not explain the Termux restriction: %s", screenView)
		}
	}
}

func TestRemoteRuntimeBlocksHostOperations(t *testing.T) {
	model := New()
	model.statusLoaded = true
	model.status.Host.Termux = false

	updated, cmd := model.toggleWorkstation()
	model = updated.(Model)
	if cmd != nil || model.busy || model.message != hostActionUnavailableMessage {
		t.Fatalf("remote start was not blocked: cmd=%v busy=%v message=%q", cmd != nil, model.busy, model.message)
	}

	updated, cmd = model.installSelectedTool()
	model = updated.(Model)
	if cmd != nil || model.installingTool != "" || model.message != hostActionUnavailableMessage {
		t.Fatalf("remote installation was not blocked: cmd=%v installing=%q message=%q", cmd != nil, model.installingTool, model.message)
	}

	updated, cmd = model.runHostOperation("setup", "setup", "--json")
	model = updated.(Model)
	if cmd != nil || model.message != hostActionUnavailableMessage {
		t.Fatalf("remote setup was not blocked: cmd=%v message=%q", cmd != nil, model.message)
	}
}

func TestRemoteHomeKeyboardKeepsStatusAndShellAvailable(t *testing.T) {
	model := New()
	model.statusLoaded = true
	model.status.Host.Termux = false

	if count := model.controlCount(); count != 2 {
		t.Fatalf("remote home control count = %d, want 2", count)
	}
	model.focus = 0
	if _, handled := model.activateFocusedControl(); !handled || model.screen != statusScreen {
		t.Fatal("remote home status control is not available")
	}
	model.navigate(homeScreen)
	model.focus = 1
	if _, handled := model.activateFocusedControl(); !handled || model.screen != shellScreen {
		t.Fatal("remote home shell control is not available")
	}
}

func TestSystemRendersFigmaSections(t *testing.T) {
	model := New()
	model.width = 40
	model.height = 40
	model.version.Version = "0.1.0"
	model.version.Channel = "stable"
	model.version.OS = "linux"
	model.version.Architecture = "arm64"
	view := ansi.Strip(model.renderSystem())
	for _, expected := range []string{"SISTEMA", "Mobdesk", "ATUALIZAÇÃO", "Verificar", "Atualizar", "DETALHES DA VERSÃO", "VERSÃO", "CANAL", "PLATAFORMA", "ÁREA AVANÇADA", "Voltar"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("system view does not contain %q: %s", expected, view)
		}
	}
	if strings.Contains(view, "Logs") {
		t.Fatalf("system view still exposes the removed Logs action: %s", view)
	}
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > contentWidth(model.width) {
			t.Fatalf("system line exceeds terminal width: %q", line)
		}
	}
}

func TestSystemRendersPersistentUpdateResult(t *testing.T) {
	model := New()
	model.screen = systemScreen
	model.width = 80
	model.height = 30

	for _, result := range []operationResult{
		{Success: true, State: "current", Message: "Mobdesk v1.0.0 já está atualizado"},
		{Success: true, State: "available", Message: "Atualização disponível: v1.0.0 → v1.1.0"},
		{Success: true, State: "updated", Message: "Mobdesk atualizado: v1.0.0 → v1.1.0"},
		{Success: false, State: "failed", Message: "rede indisponível"},
	} {
		updated, _ := model.Update(operationMessage{command: "update", result: result})
		model = updated.(Model)
		view := ansi.Strip(model.renderSystem())
		for _, expected := range []string{"RESULTADO", result.Message} {
			if !strings.Contains(view, expected) {
				t.Fatalf("system result does not contain %q: %s", expected, view)
			}
		}
	}
}

func TestSystemUpdateResultUsesFailureForProcessError(t *testing.T) {
	model := New()
	model.screen = systemScreen
	updated, _ := model.Update(operationMessage{command: "update", err: errors.New("falha de rede")})
	model = updated.(Model)

	if model.systemState != "failed" || model.systemMessage != "falha de rede" {
		t.Fatalf("system failure = state %q, message %q", model.systemState, model.systemMessage)
	}
}

func intPtr(value int) *int { return &value }

func TestHeaderStaysOnOneFullWidthLine(t *testing.T) {
	for _, width := range []int{20, 40, 80} {
		model := New()
		model.width = width
		header := model.renderHeader()
		lines := strings.Split(header, "\n")
		if len(lines) > 4 {
			t.Fatalf("header content wrapped at width %d: %q", width, header)
		}
		if lipgloss.Width(lines[0]) > width {
			t.Fatalf("header content exceeds terminal width %d: %d", width, lipgloss.Width(lines[0]))
		}
	}
}

func TestMouseClickOnCloseOpensConfirmation(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 30
	updatedClick, _ := model.Update(tea.MouseClickMsg{X: 43, Y: 0, Button: tea.MouseLeft})
	model = updatedClick.(Model)
	updated, _ := model.Update(tea.MouseReleaseMsg{X: 43, Y: 0, Button: tea.MouseLeft})
	if !updated.(Model).confirmExit {
		t.Fatal("clicking the header close button did not open confirmation")
	}
}

func TestMouseClickOnHomeStatusOpensStatus(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 30
	lines := strings.Split(model.renderScreen(), "\n")
	statusLine := -1
	for index, line := range lines {
		if strings.Contains(line, "◉") && strings.Contains(line, "Status") {
			statusLine = index
			break
		}
	}
	if statusLine < 0 {
		t.Fatal("home status card was not rendered")
	}
	updatedClick, _ := model.Update(tea.MouseClickMsg{X: 8, Y: statusLine + 4, Button: tea.MouseLeft})
	model = updatedClick.(Model)
	updated, _ := model.Update(tea.MouseReleaseMsg{X: 8, Y: statusLine + 4, Button: tea.MouseLeft})
	if updated.(Model).screen != statusScreen {
		t.Fatal("mouse click on home status card did not open status screen")
	}
}

func TestMouseClickOnHomeRightColumnOpensShell(t *testing.T) {
	model := New()
	model.width = 80
	model.height = 40
	lines := strings.Split(model.renderScreen(), "\n")
	targetLine, targetX := -1, -1
	for index, line := range lines {
		plain := ansi.Strip(line)
		if position := strings.Index(plain, "Shell Ubuntu"); position >= 0 {
			targetLine, targetX = index, utf8.RuneCountInString(plain[:position])+2
			break
		}
	}
	if targetLine < 0 {
		t.Fatal("shell card was not rendered in the right column")
	}
	// Clica no padding esquerdo do botão, não no texto, para validar o
	// contorno inteiro como área interativa.
	// A borda superior do botão também deve ser interativa.
	updatedClick, _ := model.Update(tea.MouseClickMsg{X: targetX - 2, Y: targetLine + 3, Button: tea.MouseLeft})
	model = updatedClick.(Model)
	updated, _ := model.Update(tea.MouseReleaseMsg{X: targetX - 2, Y: targetLine + 3, Button: tea.MouseLeft})
	if updated.(Model).screen != shellScreen {
		t.Fatal("mouse click on the home right-column shell card did not open shell")
	}
}

func TestMouseClickOnHomeStartButtonStartsWorkstation(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 30
	lines := strings.Split(ansi.Strip(model.renderScreen()), "\n")
	targetLine, targetX := -1, -1
	for index, line := range lines {
		if position := strings.Index(line, "Iniciar"); position >= 0 {
			targetLine, targetX = index, position+1
			break
		}
	}
	if targetLine < 0 {
		t.Fatal("home start button was not rendered")
	}

	updatedClick, _ := model.Update(tea.MouseClickMsg{X: targetX, Y: targetLine + 4, Button: tea.MouseLeft})
	model = updatedClick.(Model)
	updated, cmd := model.Update(tea.MouseReleaseMsg{X: targetX, Y: targetLine + 4, Button: tea.MouseLeft})
	model = updated.(Model)
	if cmd == nil || !model.busy || model.operation != "start" {
		t.Fatalf("clicking home start did not dispatch start: cmd=%v busy=%v operation=%q", cmd != nil, model.busy, model.operation)
	}
}

func TestMouseClickOnHomeStopButtonOpensConfirmation(t *testing.T) {
	model := New()
	model.status.SSH.Running = true
	model.width = 44
	model.height = 30
	lines := strings.Split(ansi.Strip(model.renderScreen()), "\n")
	targetLine, targetX := -1, -1
	for index, line := range lines {
		if position := strings.Index(line, "Parar"); position >= 0 {
			targetLine, targetX = index, position+1
			break
		}
	}
	if targetLine < 0 {
		t.Fatal("home stop button was not rendered")
	}

	updatedClick, _ := model.Update(tea.MouseClickMsg{X: targetX, Y: targetLine + 4, Button: tea.MouseLeft})
	model = updatedClick.(Model)
	updated, cmd := model.Update(tea.MouseReleaseMsg{X: targetX, Y: targetLine + 4, Button: tea.MouseLeft})
	model = updated.(Model)
	if cmd != nil || !model.confirmStop {
		t.Fatalf("clicking home stop did not open confirmation: cmd=%v confirm=%v", cmd != nil, model.confirmStop)
	}
}

func TestMouseClickOnSetupContinueDispatchesSetup(t *testing.T) {
	model := New()
	model.screen = setupScreen
	model.width = 44
	model.height = 30
	lines := strings.Split(ansi.Strip(model.renderScreen()), "\n")
	targetLine, targetX := -1, -1
	for index, line := range lines {
		if position := strings.Index(line, "Continuar configuração"); position >= 0 {
			targetLine, targetX = index, position+1
			break
		}
	}
	if targetLine < 0 {
		t.Fatal("setup continue button was not rendered")
	}

	updatedClick, _ := model.Update(tea.MouseClickMsg{X: targetX, Y: targetLine + 4, Button: tea.MouseLeft})
	model = updatedClick.(Model)
	updated, cmd := model.Update(tea.MouseReleaseMsg{X: targetX, Y: targetLine + 4, Button: tea.MouseLeft})
	model = updated.(Model)
	if cmd == nil || !model.busy || model.operation != "setup" {
		t.Fatalf("clicking setup continue did not dispatch setup: cmd=%v busy=%v operation=%q", cmd != nil, model.busy, model.operation)
	}
}

func TestMouseClickOnStatusBackButtonOnWideScreen(t *testing.T) {
	model := New()
	model.screen = statusScreen
	model.statusLoaded = true
	model.width = 80
	model.height = 40
	model.statusTable.SetRows(statusRows(status.SystemStatus{}, model.width))
	model.resize(model.width, model.height)
	lines := strings.Split(model.renderScreen(), "\n")
	targetLine, targetX := -1, -1
	for index, line := range lines {
		plain := ansi.Strip(line)
		if position := strings.Index(plain, "Voltar"); position >= 0 {
			targetLine, targetX = index, position+2
			break
		}
	}
	if targetLine < 0 {
		t.Fatal("status back button was not rendered")
	}
	updatedClick, _ := model.Update(tea.MouseClickMsg{X: targetX - 2, Y: targetLine + 3, Button: tea.MouseLeft})
	model = updatedClick.(Model)
	updated, _ := model.Update(tea.MouseReleaseMsg{X: targetX - 2, Y: targetLine + 3, Button: tea.MouseLeft})
	if updated.(Model).screen != homeScreen {
		t.Fatal("mouse click on wide status back button did not return home")
	}
}

func TestMouseReleaseWithNoneCompletesTouchClick(t *testing.T) {
	model := New()
	model.screen = statusScreen
	model.statusLoaded = true
	model.width = 80
	model.height = 40
	model.statusTable.SetRows(statusRows(status.SystemStatus{}, model.width))
	model.resize(model.width, model.height)
	lines := strings.Split(model.renderScreen(), "\n")
	targetLine, targetX := -1, -1
	for index, line := range lines {
		plain := ansi.Strip(line)
		if position := strings.Index(plain, "Voltar"); position >= 0 {
			targetLine, targetX = index, position+2
			break
		}
	}
	if targetLine < 0 {
		t.Fatal("status back button was not rendered")
	}
	updatedClick, _ := model.Update(tea.MouseClickMsg{X: targetX - 2, Y: targetLine + 3, Button: tea.MouseLeft})
	model = updatedClick.(Model)
	updated, _ := model.Update(tea.MouseReleaseMsg{X: targetX - 2, Y: targetLine + 3, Button: tea.MouseNone})
	if updated.(Model).screen != homeScreen {
		t.Fatal("MouseNone release did not complete the touch click")
	}
}

func TestConfirmationButtonsAreClickable(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 30
	model.confirmExit = true

	updatedClick, _ := model.Update(tea.MouseClickMsg{X: model.width/2 - 8, Y: 8, Button: tea.MouseLeft})
	model = updatedClick.(Model)
	updated, cmd := model.Update(tea.MouseReleaseMsg{X: model.width/2 - 8, Y: 8, Button: tea.MouseLeft})
	if cmd == nil {
		t.Fatal("clicking confirm did not produce a quit command")
	}
	if updated.(Model).confirmExit {
		t.Fatal("confirm modal remained open after clicking SIM")
	}

	model.confirmExit = true
	updatedClick, _ = model.Update(tea.MouseClickMsg{X: model.width/2 + 8, Y: 8, Button: tea.MouseLeft})
	model = updatedClick.(Model)
	updated, cmd = model.Update(tea.MouseReleaseMsg{X: model.width/2 + 8, Y: 8, Button: tea.MouseLeft})
	if cmd != nil {
		t.Fatal("clicking cancel unexpectedly produced a command")
	}
	if updated.(Model).confirmExit {
		t.Fatal("confirm modal remained open after clicking NÃO")
	}
}

func TestMouseWheelScrollsContent(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 10
	updated, _ := model.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	model = updated.(Model)
	if model.viewport.YOffset() == 0 {
		t.Fatal("mouse wheel did not move the viewport")
	}
}

func TestMouseDragScrollsContent(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 10
	model.viewport.SetHeight(6)
	model.viewport.SetContent(strings.Repeat("linha\n", 30))
	updatedClick, _ := model.Update(tea.MouseClickMsg{X: 2, Y: 10, Button: tea.MouseLeft})
	model = updatedClick.(Model)

	updated, _ := model.Update(tea.MouseMotionMsg{X: 2, Y: 5, Button: tea.MouseLeft})
	model = updated.(Model)
	if model.viewport.YOffset() == 0 {
		t.Fatal("dragging the mouse did not scroll the viewport")
	}

	updated, _ = model.Update(tea.MouseReleaseMsg{X: 2, Y: 5, Button: tea.MouseLeft})
	if updated.(Model).dragging {
		t.Fatal("mouse release did not stop dragging")
	}
}

func TestMouseMotionWithoutButtonCancelsStuckDrag(t *testing.T) {
	model := New()
	model.pointerDown = true
	model.dragging = true
	model.dragY = 10

	updated, _ := model.Update(tea.MouseMotionMsg{X: 4, Y: 9, Button: tea.MouseNone})
	model = updated.(Model)
	if model.pointerDown || model.dragging {
		t.Fatal("mouse motion without a pressed button did not cancel dragging")
	}
}

func TestTabFocusAndEnterOpenTools(t *testing.T) {
	model := New()
	for i := 0; i < 3; i++ {
		updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
		model = updated.(Model)
	}
	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if updated.(Model).screen != toolsScreen {
		t.Fatalf("focused home control did not open tools screen: %v", updated.(Model).screen)
	}
}

func TestHomeLayoutSupportsNarrowAndWideTerminals(t *testing.T) {
	for _, width := range []int{20, 40, 44, 80} {
		model := New()
		model.width = width
		model.height = 40
		view := model.render()
		if view == "" {
			t.Fatalf("empty view at width %d", width)
		}
	}
}

func TestMockBackendProvidesVisualScenarios(t *testing.T) {
	for _, scenario := range []string{"healthy", "degraded", "error"} {
		backend := NewMockBackend(scenario)
		message, ok := backend.StatusCmd()().(statusMessage)
		if !ok {
			t.Fatalf("scenario %q returned an unexpected status message", scenario)
		}
		if message.value.GeneratedAt.IsZero() {
			t.Fatalf("scenario %q returned status without timestamp", scenario)
		}
		if scenario == "healthy" && message.value.Overall != status.StateHealthy {
			t.Fatalf("healthy scenario overall = %q", message.value.Overall)
		}
		if scenario == "degraded" && message.value.SSH.Running {
			t.Fatalf("degraded scenario unexpectedly reports running SSH")
		}
		if scenario == "error" && message.value.Overall != status.StateError {
			t.Fatalf("error scenario overall = %q", message.value.Overall)
		}
	}
}

func TestMockBackendMutatesStateAfterInstallation(t *testing.T) {
	backend := NewMockBackend("degraded")
	message, ok := backend.OperationCmd("install", "go", "--json")().(operationMessage)
	if !ok || message.err != nil {
		t.Fatalf("mock installation failed: %+v", message)
	}
	statusMessageValue := backend.StatusCmd()().(statusMessage)
	if !toolInstalled(statusMessageValue.value, toolEntries("language")[0]) {
		t.Fatal("mock installation did not mark go as installed")
	}
}

func TestMockBackendErrorScenarioReturnsOperationError(t *testing.T) {
	backend := NewMockBackend("error")
	message, ok := backend.OperationCmd("start", "--json")().(operationMessage)
	if !ok || message.err == nil {
		t.Fatalf("mock error scenario did not return an operation error: %+v", message)
	}
}

func TestMockBackendRejectsCommandsOutsideCLI(t *testing.T) {
	backend := NewMockBackend("healthy")
	for _, command := range []string{"update-check", "setup-upgrade"} {
		message, ok := backend.OperationCmd(command, "--json")().(operationMessage)
		if !ok || message.err == nil {
			t.Fatalf("mock accepted command outside CLI contract: %q", command)
		}
	}
}

func TestHostOperationBlocksDuplicatesAndIgnoresStaleStatus(t *testing.T) {
	backend := &controlledBackend{}
	model := NewWithBackend(backend)
	model.statusLoaded = true
	model.status.Host.Termux = true

	updated, first := model.runHostOperation("install", "install", "go", "--json")
	model = updated.(Model)
	if first == nil || !model.busy || backend.operationCalls != 0 {
		t.Fatalf("first operation was not queued correctly: busy=%v calls=%d", model.busy, backend.operationCalls)
	}

	updated, duplicate := model.runHostOperation("install", "install", "python", "--json")
	model = updated.(Model)
	if duplicate != nil || model.operation != "install" {
		t.Fatalf("duplicate operation was not blocked: cmd=%v operation=%q", duplicate != nil, model.operation)
	}

	stale := status.SystemStatus{Host: status.HostStatus{Termux: true}, SSH: status.SSHStatus{Running: false}}
	updated, _ = model.Update(statusMessage{id: 1, value: stale})
	model = updated.(Model)
	if !model.busy || model.status.SSH.Running {
		t.Fatalf("stale status changed active operation state: busy=%v ssh=%v", model.busy, model.status.SSH.Running)
	}

	operation := first().(operationMessage)
	updated, refresh := model.Update(operation)
	model = updated.(Model)
	if model.busy || refresh == nil {
		t.Fatalf("operation did not finish and request fresh status: busy=%v refresh=%v", model.busy, refresh != nil)
	}

	updated, _ = model.Update(statusMessage{id: 2, value: stale})
	model = updated.(Model)
	if model.status.SSH.Running {
		t.Fatal("status requested before operation completion was accepted")
	}

	updated, _ = model.Update(refresh().(statusMessage))
	model = updated.(Model)
	if !model.status.SSH.Running || model.statusID != 3 {
		t.Fatalf("fresh status was not applied: ssh=%v statusID=%d", model.status.SSH.Running, model.statusID)
	}
}

func TestExitCancelsCancelableBackend(t *testing.T) {
	backend := &controlledBackend{}
	model := NewWithBackend(backend)
	model.confirmExit = true

	updated, cmd := model.updateConfirmation("y")
	model = updated.(Model)
	if !backend.canceled || cmd == nil || !model.closing {
		t.Fatalf("exit did not cancel backend: canceled=%v cmd=%v closing=%v", backend.canceled, cmd != nil, model.closing)
	}
}

func TestRunCommandUsesCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	message, ok := runCommand(ctx, "version")().(operationMessage)
	if !ok || !errors.Is(message.err, context.Canceled) {
		t.Fatalf("canceled command = %+v", message)
	}
}

type controlledBackend struct {
	operationCalls int
	operationArgs  []string
	canceled       bool
}

func (b *controlledBackend) StatusCmd() tea.Cmd {
	return func() tea.Msg {
		return statusMessage{value: status.SystemStatus{
			Host: status.HostStatus{Termux: true},
			SSH:  status.SSHStatus{Running: true},
		}}
	}
}

func (b *controlledBackend) OperationCmd(args ...string) tea.Cmd {
	b.operationArgs = append([]string(nil), args...)
	return func() tea.Msg {
		b.operationCalls++
		return operationMessage{command: args[0], result: operationResult{Success: true}}
	}
}

func (b *controlledBackend) ShellCmd() tea.Cmd { return nil }

func (b *controlledBackend) Cancel() { b.canceled = true }
