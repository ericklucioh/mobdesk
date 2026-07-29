package tui

import (
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/status"
)

func TestTUIScreensUseSelectedLocale(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocaleENUS, i18n.LocalePTBR} {
		t.Run(string(locale), func(t *testing.T) {
			model := NewWithBackendLocale(&controlledBackend{}, locale)
			model.width = 44
			model.height = 40
			model.statusLoaded = true
			model.status.Host.Termux = true
			model.statusTable.SetRows(statusRows(model.status, model.width, model.localizer))
			for _, view := range []string{model.renderHome(), model.renderSetup(), model.renderTools(), model.renderShell(), model.renderSystem(), model.renderStatus()} {
				if !strings.Contains(ansi.Strip(view), model.text(i18n.TUIHomeTag, nil)) && strings.Contains(ansi.Strip(view), "HOME") {
					t.Fatalf("localized view leaked English home label: %s", view)
				}
			}
			if !strings.Contains(ansi.Strip(model.renderHome()), model.text(i18n.TUIHomeStart, nil)) {
				t.Fatalf("home does not use selected locale: %s", model.renderHome())
			}
			if !strings.Contains(ansi.Strip(model.renderSystem()), model.text(i18n.TUISystemUpdate, nil)) {
				t.Fatalf("system does not use selected locale: %s", model.renderSystem())
			}
		})
	}
}

func TestTUISelectedLocaleFitsNarrowScreens(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocaleENUS, i18n.LocalePTBR} {
		model := NewWithBackendLocale(&controlledBackend{}, locale)
		model.width = 20
		model.height = 40
		model.statusLoaded = true
		model.status.Host.Termux = true
		model.statusTable.SetRows(statusRows(model.status, model.width, model.localizer))
		model.resize(model.width, model.height)
		for _, view := range []string{model.renderHome(), model.renderSetup(), model.renderTools(), model.renderShell(), model.renderSystem(), model.renderStatus()} {
			for _, line := range strings.Split(ansi.Strip(view), "\n") {
				if lipgloss.Width(line) > contentWidth(model.width) {
					t.Fatalf("%s line exceeds narrow terminal: %q", locale, line)
				}
			}
		}
	}
}

func TestTUIRemoteRestrictionUsesSelectedLocale(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocaleENUS, i18n.LocalePTBR} {
		model := NewWithBackendLocale(&controlledBackend{}, locale)
		model.statusLoaded = true
		model.status.Host.Termux = false
		model.width = 44
		for _, view := range []string{model.renderHome(), model.renderSetup(), model.renderTools(), model.renderSystem()} {
			if !strings.Contains(ansi.Strip(view), model.text(i18n.TUIHostOnlyTitle, nil)) {
				if view != model.renderHome() {
					t.Fatalf("%s restriction view omitted localized title: %s", locale, view)
				}
			}
		}
		updated, command := model.toggleWorkstation()
		model = updated.(Model)
		if command != nil || model.message != model.text(i18n.TUIHostRestriction, nil) {
			t.Fatalf("%s remote operation was not localized or blocked", locale)
		}
	}
}

func TestPopupKeyboardAndMouseConfirmationUseSelectedLocale(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocaleENUS, i18n.LocalePTBR} {
		t.Run(string(locale), func(t *testing.T) {
			backend := &controlledBackend{}
			model := localizedPopupTestModel(backend, locale)
			model.popupFocus = 1
			updated, command := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
			model = updated.(Model)
			if command != nil || !model.popupConfirm || !strings.Contains(ansi.Strip(model.renderScreen()), model.text(i18n.TUIPopupConfirm, nil)) {
				t.Fatalf("keyboard confirmation missing for %s", locale)
			}
			updated, command = model.Update(tea.KeyPressMsg(tea.Key{Text: "n"}))
			model = updated.(Model)
			if command != nil || model.popupConfirm {
				t.Fatalf("keyboard cancellation failed for %s", locale)
			}

			model.popupFocus = 1
			updated, _ = model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
			model = updated.(Model)
			updated, _ = model.Update(tea.KeyPressMsg(tea.Key{Text: "n"}))
			model = updated.(Model)
			lines := strings.Split(ansi.Strip(model.renderScreen()), "\n")
			actionLabel := "[ " + model.text(i18n.TUIPopupUninstall, nil) + " ]"
			actionLine := findLine(lines, actionLabel)
			if actionLine < 0 {
				t.Fatal("localized uninstall action is missing")
			}
			actionX := utf8.RuneCountInString(lines[actionLine][:strings.Index(lines[actionLine], actionLabel)]) + 1
			updated, _ = model.Update(tea.MouseClickMsg{X: actionX, Y: actionLine + 3, Button: tea.MouseLeft})
			model = updated.(Model)
			updated, command = model.Update(tea.MouseReleaseMsg{X: actionX, Y: actionLine + 3, Button: tea.MouseLeft})
			model = updated.(Model)
			if command != nil || !model.popupConfirm {
				t.Fatalf("mouse did not open localized confirmation for %s", locale)
			}
			confirmationLines := strings.Split(ansi.Strip(model.renderScreen()), "\n")
			yes := model.text(i18n.TUIConfirmationYes, nil)
			yesLine := findLine(confirmationLines, yes)
			yesX := utf8.RuneCountInString(confirmationLines[yesLine][:strings.Index(confirmationLines[yesLine], yes)]) + 1
			updated, _ = model.Update(tea.MouseClickMsg{X: yesX, Y: yesLine + 3, Button: tea.MouseLeft})
			model = updated.(Model)
			updated, command = model.Update(tea.MouseReleaseMsg{X: yesX, Y: yesLine + 3, Button: tea.MouseLeft})
			if command == nil || backend.operationArgs[0] != "uninstall" {
				t.Fatalf("mouse did not dispatch localized destructive action for %s", locale)
			}
		})
	}
}

func TestMockBackendUsesSelectedLocale(t *testing.T) {
	for _, locale := range []i18n.Locale{i18n.LocaleENUS, i18n.LocalePTBR} {
		message := NewMockBackendLocale("healthy", locale).OperationCmd("setup")().(operationMessage)
		if message.result.Message != i18n.New(locale).Text(i18n.TUIOperationCompleted, nil) {
			t.Fatalf("mock message ignored locale %s: %q", locale, message.result.Message)
		}
	}
}

func localizedPopupTestModel(backend *controlledBackend, locale i18n.Locale) Model {
	model := NewWithBackendLocale(backend, locale)
	model.screen = toolsScreen
	model.statusLoaded = true
	model.status.Host.Termux = true
	model.status.Installations = []status.InstallationStatus{{Name: "neovim", Kind: "editor", Package: "neovim", Executable: "nvim", State: "installed", Source: "mobdesk", Managed: true, Version: "0.11"}}
	model.openAppPopup(toolIndex("neovim"))
	model.width = 44
	model.height = 30
	return model
}

func findLine(lines []string, value string) int {
	for index, line := range lines {
		if strings.Contains(line, value) {
			return index
		}
	}
	return -1
}
