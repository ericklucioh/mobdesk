package tui

import (
	"strings"
	"testing"
)

func TestScreensRenderAtPhoneWidth(t *testing.T) {
	model := NewWithBackend(NewMockBackend("healthy"))
	model.status = mockStatus("healthy")
	model.statusLoaded = true
	model.resize(32, 18)

	tests := []struct {
		name   string
		screen screen
	}{
		{name: "home", screen: homeScreen},
		{name: "status", screen: statusScreen},
		{name: "setup", screen: setupScreen},
		{name: "tools", screen: toolsScreen},
		{name: "shell", screen: shellScreen},
		{name: "system", screen: systemScreen},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model.screen = test.screen
			if rendered := model.renderScreen(); strings.TrimSpace(rendered) == "" {
				t.Fatalf("%s did not render", test.name)
			}
		})
	}
}

func TestToolsRenderCurrentCatalogAndInstallingState(t *testing.T) {
	model := NewWithBackend(NewMockBackend("healthy"))
	model.status = mockStatus("healthy")
	model.statusLoaded = true
	model.screen = toolsScreen
	model.resize(80, 60)
	model.installingTool = "node"

	items := toolListItemsLocalized(model.status, model.installingTool, model.localizer)
	if len(items) != len(toolEntriesLocalized("", model.localizer)) {
		t.Fatalf("tools list omitted catalog profiles: %d", len(items))
	}
	for _, item := range items {
		if item.entry.profile.Name == "node" && !item.installing {
			t.Fatal("installing state was not rendered for node")
		}
	}
}
