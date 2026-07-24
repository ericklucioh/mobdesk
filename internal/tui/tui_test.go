package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func TestHomeRendersWorkstationCard(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 50
	model.height = 30
	view := model.renderScreen()
	for _, expected := range []string{"INÍCIO", "Workstation SSH", "Status: desativado", "Apps e linguagens"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("home view does not contain %q: %s", expected, view)
		}
	}
}

func TestToolsUseCurrentCatalog(t *testing.T) {
	model := New()
	model.width = 44
	model.height = 50
	view := model.renderToolsBubbles()
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
