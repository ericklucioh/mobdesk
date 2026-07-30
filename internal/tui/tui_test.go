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
	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/status"
)

func TestHomeRendersWorkstationCard(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 50
	model.height = 30
	view := model.renderScreen()
	for _, expected := range []string{model.text(i18n.TUIHomeTag, nil), model.text(i18n.TUIHomeWorkstationTitle, nil), model.text(i18n.TUIHomeStatusLabel, nil), model.text(i18n.TUIStateStopped, nil), model.text(i18n.TUIHomeAppsTitle, nil)} {
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
	if !strings.Contains(view, "┌") || !strings.Contains(view, model.text(i18n.TUIToolStateInstall, nil)) {
		t.Fatalf("tools view does not render bordered install items: %s", view)
	}
	lines := strings.Split(view, "\n")
	goLine := -1
	for index, line := range lines {
		if strings.Contains(line, "go") {
			goLine = index
			if !strings.Contains(line, model.text(i18n.TUIToolStateInstall, nil)) {
				t.Fatalf("tool state is not on the App line: %q", line)
			}
			break
		}
	}
	if goLine < 0 || goLine+1 >= len(lines) || !strings.Contains(lines[goLine+1], model.localizer.Text(i18n.AppGoDescription, nil)) {
		t.Fatalf("tool phrase is not directly below App: %q", lines)
	}
	model.statusLoaded = true
	model.status.Host.Termux = true
	model.status.Installations = []status.InstallationStatus{{Name: "go", Kind: "language", State: "installed"}}
	if !toolInstalled(model.status, toolEntries("language")[0]) {
		t.Fatal("installed go was not recognized")
	}
	view = ansi.Strip(model.renderTools())
	if !strings.Contains(view, model.text(i18n.TUIToolStateInstalled, nil)) {
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
	if view := ansi.Strip(model.renderTools()); !strings.Contains(view, model.text(i18n.TUIToolStateInstalling, nil)) {
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
	if model.installingTool != "" || !strings.Contains(ansi.Strip(model.renderTools()), model.text(i18n.TUIToolStateInstalled, nil)) {
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
	for _, expected := range []string{model.text(i18n.TUIOperationStart, nil), model.text(i18n.TUIOperationRunning, nil), model.text(i18n.TUIOperationWait, nil)} {
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
		t.Fatalf("tools view does not render last catalog item %q", entries[last].profile.Name)
	}

	updated, cmd := model.installSelectedTool()
	model = updated.(Model)
	if cmd == nil || model.installingTool != entries[last].profile.Name {
		t.Fatalf("last catalog item was not installable: cmd=%v installing=%q", cmd != nil, model.installingTool)
	}
}

func TestReleaseToolRowsOpenDetailsBeforeInstall(t *testing.T) {
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
				if entry.profile.Name == name {
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
			if command != nil || !model.appPopupOpen || model.installingTool != "" {
				t.Fatalf("row did not open details for %s: command=%v popup=%v installing=%q", name, command != nil, model.appPopupOpen, model.installingTool)
			}
			popupLines := strings.Split(ansi.Strip(model.renderScreen()), "\n")
			installLine := -1
			for candidate, line := range popupLines {
				if strings.Contains(line, "[ "+model.text(i18n.TUIPopupInstall, nil)+" ]") {
					installLine = candidate
					break
				}
			}
			if installLine < 0 {
				t.Fatalf("install action was not rendered for %s: %s", name, model.renderScreen())
			}
			actionLabel := "[ " + model.text(i18n.TUIPopupInstall, nil) + " ]"
			actionX := utf8.RuneCountInString(popupLines[installLine][:strings.Index(popupLines[installLine], actionLabel)]) + 1
			updated, _ = model.Update(tea.MouseClickMsg{X: actionX, Y: installLine + 4, Button: tea.MouseLeft})
			model = updated.(Model)
			updated, command = model.Update(tea.MouseReleaseMsg{X: actionX, Y: installLine + 4, Button: tea.MouseLeft})
			model = updated.(Model)
			if command == nil || model.installingTool != name {
				t.Fatalf("popup did not start %s: command=%v installing=%q", name, command != nil, model.installingTool)
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
	for _, expected := range []string{model.text(i18n.TUISetupTag, nil), model.text(i18n.TUISetupTitle, nil), model.text(i18n.TUISetupDirectories, nil), model.text(i18n.TUISetupWorkspace, nil), model.text(i18n.TUISetupAdvanced, nil)} {
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

func TestSetupRendersPersistedPhaseStates(t *testing.T) {
	model := New()
	model.width = 40
	model.height = 40
	model.statusLoaded = true
	model.status.Host.Termux = true
	model.status.Setup.Phases = map[string]string{
		"directories":         "done",
		"packages-installed":  "done",
		"ubuntu-installed":    "pending",
		"workspace-created":   "pending",
		"password-configured": "pending",
		"ssh-configured":      "pending",
		"shell-configured":    "pending",
		"launcher-installed":  "pending",
	}

	view := ansi.Strip(model.renderSetup())
	if strings.Count(view, "✓") != 2 {
		t.Fatalf("setup rendered incorrect completed markers: %s", view)
	}
	if !strings.Contains(view, "warning") || !strings.Contains(view, "○") {
		t.Fatalf("setup did not render pending phase state: %s", view)
	}
}

func TestShellRendersLargeTouchActions(t *testing.T) {
	model := New()
	model.width = 40
	model.height = 30
	view := model.renderShell()
	for _, expected := range []string{model.text(i18n.TUIShellOpen, nil), "Suspend the TUI", model.text(i18n.TUIShellBack, nil), "Return to the main"} {
		if !strings.Contains(ansi.Strip(view), expected) {
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
	refreshLabel := model.text(i18n.TUIStatusRefresh, nil)
	if contentWidth(model.width) < 28 {
		refreshLabel = model.text(i18n.TUIStatusRefreshShort, nil)
	}
	for _, expected := range []string{model.text(i18n.TUIStatusTag, nil), model.text(i18n.TUIStatusTitle, nil), model.text(i18n.TUIStatusHost, nil), model.text(i18n.TUIStatusDetails, nil), refreshLabel} {
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
	for _, expected := range []string{model.text(i18n.TUIHomeRemoteTitle, nil), model.text(i18n.TUIHomeStatusTitle, nil), model.text(i18n.TUIHomeShellTitle, nil)} {
		if !strings.Contains(home, expected) {
			t.Fatalf("remote home does not contain %q: %s", expected, home)
		}
	}
	if strings.Contains(home, model.text(i18n.TUIHomeStart, nil)) || strings.Contains(home, model.text(i18n.TUIHomeSetupTitle, nil)) {
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
	for _, expected := range []string{model.text(i18n.TUISystemTag, nil), model.text(i18n.TUISystemTitle, nil), model.text(i18n.TUISystemUpdate, nil), model.text(i18n.TUISystemCheck, nil), model.text(i18n.TUISystemUpdate, nil), model.text(i18n.TUISystemVersion, nil), model.text(i18n.TUISystemVersion, nil), model.text(i18n.TUISystemChannel, nil), model.text(i18n.TUISystemPlatform, nil), model.text(i18n.TUISystemAdvanced, nil), model.text(i18n.TUIStatusBack, nil)} {
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
		{Success: true, State: "current", Message: model.text(i18n.OutputUpdateCurrent, map[string]any{"Version": "v1.0.0"})},
		{Success: true, State: "available", Message: model.text(i18n.OutputUpdateAvailable, map[string]any{"Current": "v1.0.0", "Latest": "v1.1.0"})},
		{Success: true, State: "updated", Message: model.text(i18n.OutputUpdateUpdated, map[string]any{"Current": "v1.0.0", "Latest": "v1.1.0"})},
		{Success: false, State: "failed", Message: model.text(i18n.ErrorOperationFailed, map[string]any{"Detail": "network unavailable"})},
	} {
		updated, _ := model.Update(operationMessage{command: "update", result: result})
		model = updated.(Model)
		view := ansi.Strip(model.renderSystem())
		for _, expected := range []string{model.text(i18n.TUISystemResult, nil), result.Message} {
			if !strings.Contains(view, expected) {
				t.Fatalf("system result does not contain %q: %s", expected, view)
			}
		}
	}
}

func TestSystemUpdateResultUsesFailureForProcessError(t *testing.T) {
	model := New()
	model.screen = systemScreen
	updated, _ := model.Update(operationMessage{command: "update", err: errors.New("network failure")})
	model = updated.(Model)

	if model.systemState != "failed" || model.systemMessage != model.text(i18n.ErrorOperationFailed, map[string]any{"Detail": "network failure"}) {
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
	header := ansi.Strip(model.renderHeader())
	closeLabel := "[ " + model.text(i18n.TUIHeaderClose, nil) + " ]"
	closeLine := strings.Split(header, "\n")[1]
	closeStart := strings.Index(closeLine, closeLabel)
	closeX := utf8.RuneCountInString(closeLine[:closeStart]) + 1
	updatedClick, _ := model.Update(tea.MouseClickMsg{X: closeX, Y: 1, Button: tea.MouseLeft})
	model = updatedClick.(Model)
	updated, _ := model.Update(tea.MouseReleaseMsg{X: closeX, Y: 1, Button: tea.MouseLeft})
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
		if position := strings.Index(plain, model.text(i18n.TUIHomeShellTitle, nil)); position >= 0 {
			targetLine, targetX = index, utf8.RuneCountInString(plain[:position])+2
			break
		}
	}
	if targetLine < 0 {
		t.Fatal("shell card was not rendered in the right column")
	}
	// Click the button's left padding, not its text, to validate the entire
	// outline as an interactive area. The top border must also be interactive.
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
		if position := strings.Index(line, model.text(i18n.TUIHomeStart, nil)); position >= 0 {
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
		if position := strings.Index(line, model.text(i18n.TUIHomeStop, nil)); position >= 0 {
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
		if position := strings.Index(line, model.text(i18n.TUISetupContinue, nil)); position >= 0 {
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
		if position := strings.Index(plain, model.text(i18n.TUIStatusBack, nil)); position >= 0 {
			targetLine, targetX = index, utf8.RuneCountInString(plain[:position])+2
			break
		}
	}
	if targetLine < 0 {
		t.Fatal("status back button was not rendered")
	}
	plainBack := ansi.Strip(lines[targetLine])
	byteStart := strings.Index(strings.ToLower(plainBack), strings.ToLower(model.text(i18n.TUIStatusBack, nil)))
	targetX = utf8.RuneCountInString(plainBack[:byteStart]) + 2
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
		if position := strings.Index(plain, model.text(i18n.TUIStatusBack, nil)); position >= 0 {
			targetLine, targetX = index, utf8.RuneCountInString(plain[:position])+2
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
		t.Fatal("confirm modal remained open after clicking NO")
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

func TestHomeTabFocusCoversOnlyRenderedControls(t *testing.T) {
	model := New()
	if count := model.controlCount(); count != 6 {
		t.Fatalf("home control count = %d, want 6", count)
	}
	for index := 0; index < model.controlCount(); index++ {
		updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
		model = updated.(Model)
		want := (index + 1) % 6
		if model.focus != want {
			t.Fatalf("tab %d moved focus to %d, want %d", index+1, model.focus, want)
		}
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

func TestToolsPopupOpensWithKeyboardAndClosesWithEscape(t *testing.T) {
	model := NewWithBackend(&controlledBackend{})
	model.screen = toolsScreen
	model.width = 44
	model.height = 30
	updated, command := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	model = updated.(Model)
	if command != nil || !model.appPopupOpen || model.popupAppIndex != 0 {
		t.Fatalf("Enter did not open app popup: command=%v popup=%v index=%d", command != nil, model.appPopupOpen, model.popupAppIndex)
	}
	updated, command = model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	model = updated.(Model)
	if command != nil || model.appPopupOpen {
		t.Fatalf("Escape did not close app popup: command=%v popup=%v", command != nil, model.appPopupOpen)
	}
}

func TestPopupInstallAndConfigActionsUseCLIContract(t *testing.T) {
	backend := &controlledBackend{}
	model := popupTestModel(backend, "neovim", "mobdesk", true, status.ConfigStateNotApplied)
	for range 2 {
		updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
		model = updated.(Model)
	}
	updated, command := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	model = updated.(Model)
	if command == nil || backend.operationArgs[0] != "config" || backend.operationArgs[1] != "apply" || backend.operationArgs[2] != "neovim" {
		t.Fatalf("config apply did not use CLI contract: command=%v args=%v", command != nil, backend.operationArgs)
	}

	model = popupTestModel(backend, "neovim", "mobdesk", true, status.ConfigStateApplied)
	for range 3 {
		updated, _ = model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
		model = updated.(Model)
	}
	updated, command = model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	model = updated.(Model)
	if command != nil || !model.popupConfirm {
		t.Fatalf("config removal did not request confirmation: command=%v confirm=%v", command != nil, model.popupConfirm)
	}
	updated, command = model.Update(tea.KeyPressMsg(tea.Key{Text: "y"}))
	model = updated.(Model)
	if command == nil || backend.operationArgs[0] != "config" || backend.operationArgs[1] != "remove" || backend.operationArgs[2] != "neovim" {
		t.Fatalf("config remove did not use CLI contract: command=%v args=%v", command != nil, backend.operationArgs)
	}
}

func TestPopupUninstallRequiresKeyboardAndMouseConfirmation(t *testing.T) {
	backend := &controlledBackend{}
	model := popupTestModel(backend, "neovim", "mobdesk", true, status.ConfigStateNotApplied)
	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	model = updated.(Model)
	updated, command := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	model = updated.(Model)
	if command != nil || !model.popupConfirm {
		t.Fatalf("uninstall did not request confirmation: command=%v confirm=%v", command != nil, model.popupConfirm)
	}
	updated, command = model.Update(tea.KeyPressMsg(tea.Key{Text: "n"}))
	model = updated.(Model)
	if command != nil || model.popupConfirm {
		t.Fatal("N did not cancel uninstall confirmation")
	}

	lines := strings.Split(ansi.Strip(model.renderScreen()), "\n")
	actionLine := -1
	for index, line := range lines {
		if strings.Contains(line, "[ "+model.text(i18n.TUIPopupUninstall, nil)+" ]") {
			actionLine = index
			break
		}
	}
	if actionLine < 0 {
		t.Fatalf("uninstall action missing from popup: %s", model.renderScreen())
	}
	actionLabel := "[ " + model.text(i18n.TUIPopupUninstall, nil) + " ]"
	actionX := utf8.RuneCountInString(lines[actionLine][:strings.Index(lines[actionLine], actionLabel)]) + 1
	updated, _ = model.Update(tea.MouseClickMsg{X: actionX, Y: actionLine + 4, Button: tea.MouseLeft})
	model = updated.(Model)
	updated, command = model.Update(tea.MouseReleaseMsg{X: actionX, Y: actionLine + 4, Button: tea.MouseLeft})
	model = updated.(Model)
	if command != nil || !model.popupConfirm {
		t.Fatalf("mouse uninstall did not request confirmation: command=%v confirm=%v", command != nil, model.popupConfirm)
	}
	lines = strings.Split(ansi.Strip(model.renderScreen()), "\n")
	confirmLine := -1
	for index, line := range lines {
		if strings.Contains(line, model.text(i18n.TUIConfirmationYes, nil)) {
			confirmLine = index
			break
		}
	}
	if confirmLine < 0 {
		t.Fatalf("confirmation buttons missing from popup: %s", model.renderScreen())
	}
	confirmLabel := model.text(i18n.TUIConfirmationYes, nil)
	confirmX := utf8.RuneCountInString(lines[confirmLine][:strings.Index(lines[confirmLine], confirmLabel)]) + 1
	updated, _ = model.Update(tea.MouseClickMsg{X: confirmX, Y: confirmLine + 4, Button: tea.MouseLeft})
	model = updated.(Model)
	updated, command = model.Update(tea.MouseReleaseMsg{X: confirmX, Y: confirmLine + 4, Button: tea.MouseLeft})
	model = updated.(Model)
	if command == nil || backend.operationArgs[0] != "uninstall" || backend.operationArgs[1] != "neovim" {
		t.Fatalf("mouse confirmation did not dispatch uninstall: command=%v args=%v", command != nil, backend.operationArgs)
	}
}

func TestPopupBlocksDetectedAppAndRemoteRuntime(t *testing.T) {
	detected := popupTestModel(&controlledBackend{}, "neovim", "detected", false, status.ConfigStateNotApplied)
	detected.popupFocus = 1
	updated, command := detected.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	detected = updated.(Model)
	if command != nil || !strings.Contains(detected.popupMessage, detected.text(i18n.TUIPopupDetectedReason, nil)) {
		t.Fatalf("detected app was not blocked: command=%v message=%q", command != nil, detected.popupMessage)
	}

	remote := popupTestModel(&controlledBackend{}, "neovim", "mobdesk", true, status.ConfigStateNotApplied)
	remote.status.Host.Termux = false
	remote.popupFocus = 0
	updated, command = remote.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	remote = updated.(Model)
	if command != nil || !strings.Contains(remote.popupMessage, remote.text(i18n.TUIHostRestriction, nil)) {
		t.Fatalf("remote popup did not explain restriction: command=%v message=%q", command != nil, remote.popupMessage)
	}
}

func TestPopupFitsNarrowTerminal(t *testing.T) {
	model := popupTestModel(&controlledBackend{}, "neovim", "mobdesk", true, status.ConfigStateApplied)
	model.width = 20
	model.height = 40
	for _, line := range strings.Split(ansi.Strip(model.renderScreen()), "\n") {
		if lipgloss.Width(line) > contentWidth(model.width) {
			t.Fatalf("popup line exceeds narrow terminal: %q", line)
		}
	}
}

func popupTestModel(backend *controlledBackend, name, source string, managed bool, configState status.ConfigState) Model {
	model := NewWithBackend(backend)
	model.screen = toolsScreen
	model.statusLoaded = true
	model.status.Host.Termux = true
	model.status.Installations = []status.InstallationStatus{{Name: name, Kind: "editor", Package: "neovim", Executable: "nvim", State: "installed", Source: source, Managed: managed, Version: "0.11"}}
	model.status.Configurations = []status.ConfigurationStatus{{App: name, Profile: "lazyvim", State: configState, ManagedPaths: []string{"/root/.config/nvim"}}}
	model.openAppPopup(toolIndex(name))
	model.width = 44
	model.height = 30
	return model
}

func toolIndex(name string) int {
	for index, entry := range toolEntries("") {
		if entry.profile.Name == name {
			return index
		}
	}
	return -1
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
