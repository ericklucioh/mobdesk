package tui

import "testing"

func TestToolRowHitRegionIncludesRenderedCard(t *testing.T) {
	lines := []string{
		"┌────────────────┐",
		"│ git Installed  │",
		"│ source control │",
		"└────────────────┘",
	}
	for _, index := range []int{0, 1, 2, 3} {
		if !toolRowContainsAt(lines, index, 0, "git", 18) {
			t.Fatalf("card line %d is not clickable", index)
		}
	}
}

func TestPopupButtonHitRegionIncludesPadding(t *testing.T) {
	model := NewWithBackend(NewMockBackend("healthy"))
	model.width = 80
	model.openAppPopup(toolIndex(t, model, "sqlite"))
	action := model.popupActions()[0]
	label := popupActionLabelLocalized(action, contentWidth(model.width), model.localizer)
	lines := []string{"", "     " + label + "  ", ""}
	if _, ok := model.popupActionAt(lines, 1, 3); !ok {
		t.Fatalf("button left padding is not clickable: %q", lines[1])
	}
}
