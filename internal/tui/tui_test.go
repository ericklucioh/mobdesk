package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/ericklucioh/mobdesk/internal/status"
)

func TestScreensRenderAtPhoneWidth(t *testing.T) {
	m := NewWithBackend(NewMockBackend("healthy"))
	m.status = mockStatus("healthy")
	m.statusLoaded = true
	m.resize(32, 18)
	for name, current := range map[string]screen{
		"home": homeScreen, "status": statusScreen, "setup": setupScreen, "tools": toolsScreen, "shell": shellScreen, "system": systemScreen,
	} {
		t.Run(name, func(t *testing.T) {
			m.screen = current
			if rendered := m.renderScreen(); strings.TrimSpace(rendered) == "" {
				t.Fatalf("%s did not render", name)
			}
		})
	}
}

func TestToolsRenderCurrentCatalogAndInstallingState(t *testing.T) {
	m := NewWithBackend(NewMockBackend("healthy"))
	m.status = mockStatus("healthy")
	m.statusLoaded = true
	m.screen = toolsScreen
	m.resize(80, 60)
	m.installingTool = "node"
	items := toolListItemsLocalized(m.status, m.installingTool, m.localizer)
	if len(items) != len(toolEntriesLocalized("", m.localizer)) {
		t.Fatalf("tools list omitted catalog profiles: %d", len(items))
	}
	for _, item := range items {
		if item.entry.profile.Name == "node" && !item.installing {
			t.Fatal("installing state was not rendered for node")
		}
	}
}

func TestKeyboardNavigationCoversRenderedScreenControls(t *testing.T) {
	for name, target := range map[string]screen{"setup": setupScreen, "status": statusScreen, "tools": toolsScreen, "shell": shellScreen, "system": systemScreen} {
		t.Run("home to "+name, func(t *testing.T) {
			m := NewWithBackend(NewMockBackend("healthy"))
			m.statusLoaded = true
			m.status = mockStatus("healthy")
			m.focus = map[screen]int{setupScreen: 1, statusScreen: 2, toolsScreen: 3, shellScreen: 4, systemScreen: 5}[target]
			if _, handled := m.activateFocusedControl(); !handled || m.screen != target {
				t.Fatalf("home focus %d did not open %s", m.focus, name)
			}
		})
	}

	for name, current := range map[string]screen{"status": statusScreen, "shell": shellScreen, "system": systemScreen} {
		t.Run(name+" back", func(t *testing.T) {
			m := NewWithBackend(NewMockBackend("healthy"))
			m.screen = current
			m.focus = map[screen]int{statusScreen: 1, shellScreen: 1, systemScreen: 2}[current]
			if _, handled := m.activateFocusedControl(); !handled || m.screen != homeScreen {
				t.Fatalf("%s keyboard back did not open home", name)
			}
		})
	}
}

func TestStatusRefreshReleasesBusyState(t *testing.T) {
	m := NewWithBackend(NewMockBackend("healthy"))
	m.busy, m.refreshing, m.installingTool = true, true, "node"
	m.statusID = 4
	updated, _ := m.Update(statusMessage{id: 4, value: mockStatus("healthy")})
	value := updated.(Model)
	if value.busy || value.refreshing || value.installingTool != "" {
		t.Fatalf("status refresh did not release operation state: busy=%t refreshing=%t tool=%q", value.busy, value.refreshing, value.installingTool)
	}
}

func TestStatusRefreshFailureReleasesBusyState(t *testing.T) {
	m := NewWithBackend(NewMockBackend("healthy"))
	m.busy, m.refreshing, m.installingTool = true, true, "node"
	m.statusID = 4
	updated, _ := m.Update(statusMessage{id: 4, err: fmt.Errorf("status unavailable")})
	value := updated.(Model)
	if value.busy || value.refreshing || value.installingTool != "" || value.message == "" {
		t.Fatalf("status refresh failure left operation blocked: %+v", value)
	}
}

func TestMockStatusMarksEveryCatalogProfileInstalled(t *testing.T) {
	value := mockStatus("healthy")
	for _, entry := range toolEntriesLocalized("", NewMockBackend("healthy").(*mockBackend).localizer) {
		if !toolInstalled(value, entry) {
			t.Fatalf("mock did not mark %q as installed: %+v", entry.profile.Name, value.Installations)
		}
	}
}

func TestPopupOffersManagedPartialUninstall(t *testing.T) {
	m := NewWithBackend(NewMockBackend("healthy"))
	index := toolIndex(t, m, "node")
	m.status = status.SystemStatus{
		Host: status.HostStatus{Termux: true},
		Installations: []status.InstallationStatus{{
			Name: "node", Package: "nodejs", Executable: "node", State: "partial", Source: "mobdesk", Managed: true,
		}},
	}
	m.openAppPopup(index)
	if !popupActionEnabled(m.popupActions(), "uninstall") {
		t.Fatalf("partial managed profile cannot be removed: %+v", m.popupActions())
	}
}

func TestPopupDisablesMutationsDuringStatusRefresh(t *testing.T) {
	m := NewWithBackend(NewMockBackend("healthy"))
	m.busy, m.refreshing = true, true
	m.openAppPopup(toolIndex(t, m, "git"))
	for _, action := range m.popupActions() {
		if action.ID == "close" {
			continue
		}
		if action.Enabled {
			t.Fatalf("busy popup action %q remained enabled", action.ID)
		}
	}
}

func TestToolRowHitRegionIncludesRenderedCard(t *testing.T) {
	lines := []string{"┌────────────────┐", "│ git Installed  │", "│ source control │", "└────────────────┘"}
	for _, index := range []int{0, 1, 2, 3} {
		if !toolRowContainsAt(lines, index, 0, "git", 18) {
			t.Fatalf("card line %d is not clickable", index)
		}
	}
}

func TestPopupButtonHitRegionIncludesPadding(t *testing.T) {
	m := NewWithBackend(NewMockBackend("healthy"))
	m.width = 80
	m.openAppPopup(toolIndex(t, m, "git"))
	action := m.popupActions()[0]
	label := popupActionLabelLocalized(action, contentWidth(m.width), m.localizer)
	lines := []string{"", "     " + label + "  ", ""}
	if _, ok := m.popupActionAt(lines, 1, 3); !ok {
		t.Fatalf("button left padding is not clickable: %q", lines[1])
	}
}

func TestEscapeReturnsStatusSystemAndShellToHome(t *testing.T) {
	for name, current := range map[string]screen{"status": statusScreen, "system": systemScreen, "shell": shellScreen} {
		t.Run(name, func(t *testing.T) {
			m := NewWithBackend(NewMockBackend("healthy"))
			m.screen = current
			m.history = []screen{toolsScreen}
			updated, _ := m.updateKey(tea.KeyPressMsg{Code: tea.KeyEscape})
			value := updated.(Model)
			if value.screen != homeScreen || len(value.history) != 0 {
				t.Fatalf("escape did not return home from %v: %+v", current, value.history)
			}
		})
	}
}

func toolIndex(t *testing.T, m Model, name string) int {
	t.Helper()
	for index, entry := range toolEntriesLocalized("", m.localizer) {
		if entry.profile.Name == name {
			return index
		}
	}
	t.Fatalf("missing tool %q", name)
	return 0
}

func popupActionEnabled(actions []popupAction, id string) bool {
	for _, action := range actions {
		if action.ID == id {
			return action.Enabled
		}
	}
	return false
}
