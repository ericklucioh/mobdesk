package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/ericklucioh/mobdesk/internal/i18n"
)

func (m Model) text(id i18n.MessageID, data any) string { return m.localizer.Text(id, data) }

func (m Model) localizedError(err error) string {
	if err == nil {
		return ""
	}
	if i18n.ErrorMessageID(err) != "" {
		return m.localizer.Error(err)
	}
	return m.text(i18n.ErrorOperationFailed, map[string]any{"Detail": err.Error()})
}

func renderPage(tag, title, body string) string {
	return tagStyle.Render(tag) + "\n" + titleStyle.Render(title) + "\n" + bodyStyle.Render(body)
}
func available(value bool) string {
	return availableLocalized(value, i18n.New(i18n.LocaleENUS))
}
func availableLocalized(value bool, localizer i18n.Localizer) string {
	if value {
		return localizer.Text(i18n.TUIStatusAvailable, nil)
	}
	return localizer.Text(i18n.TUIStatusInactive, nil)
}
func yesNo(value bool) string {
	return yesNoLocalized(value, i18n.New(i18n.LocaleENUS))
}
func yesNoLocalized(value bool, localizer i18n.Localizer) string {
	if value {
		return localizer.Text(i18n.TUIStatusYes, nil)
	}
	return localizer.Text(i18n.TUIStatusNo, nil)
}
func formatBytes(value int64) string {
	if value < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(value)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(value)/(1024*1024*1024))
}
func nearLine(lines []string, index int, text string) bool {
	return index >= 0 && index < len(lines) && strings.Contains(strings.ToLower(lines[index]), strings.ToLower(text))
}
func blockContainsAt(lines []string, index, x int, text string) bool {
	return blockContainsAtVertical(lines, index, x, text, 1)
}

func touchBlockContainsAt(lines []string, index, x int, text string) bool {
	return blockContainsAtVertical(lines, index, x, text, 2)
}

func blockContainsAtVertical(lines []string, index, x int, text string, verticalPadding int) bool {
	for position, line := range lines {
		plain := ansi.Strip(line)
		plainRunes := []rune(strings.ToLower(plain))
		textRunes := []rune(strings.ToLower(text))
		start := runeSliceIndex(plainRunes, textRunes)
		if start < 0 || index < position-verticalPadding || index > position+verticalPadding {
			continue
		}
		first := start
		last := first + len(textRunes) - 1
		if verticalPadding > 1 {
			// Large actions on mobile use the full rendered row as the touch target.
			lineWidth := utf8.RuneCountInString(plain)
			return x >= max(0, first-2) && x <= max(last+2, lineWidth-1)
		}
		return x >= first-2 && x <= last+2
	}
	return false
}

func runeSliceIndex(value, needle []rune) int {
	if len(needle) == 0 || len(needle) > len(value) {
		return -1
	}
	for index := 0; index <= len(value)-len(needle); index++ {
		match := true
		for offset := range needle {
			if value[index+offset] != needle[offset] {
				match = false
				break
			}
		}
		if match {
			return index
		}
	}
	return -1
}

// blockContainsAtAny is used when a screen repeats a label, such as "Status"
// in the Home card and workstation summary. It tests every occurrence before
// giving up while preserving horizontal target validation.
func blockContainsAtAny(lines []string, index, x int, text string) bool {
	for position, line := range lines {
		plain := ansi.Strip(line)
		start := strings.Index(strings.ToLower(plain), strings.ToLower(text))
		if start < 0 || index < position-1 || index > position+1 {
			continue
		}
		first := utf8.RuneCountInString(plain[:start])
		last := first + utf8.RuneCountInString(text) - 1
		if x >= first-2 && x <= last+2 {
			return true
		}
	}
	return false
}

func renderedLabelAt(lines []string, index, x int, text string, verticalPadding int) bool {
	return blockContainsAtVertical(lines, index, x, text, verticalPadding)
}

func renderedHeaderLabelAt(lines []string, index, x int, text string) bool {
	if renderedLabelAt(lines, index, x, text, 1) {
		return true
	}
	for position, line := range lines {
		if position < index-1 || position > index+1 {
			continue
		}
		plain := []rune(ansi.Strip(line))
		needle := []rune(strings.ToLower(text))
		lower := []rune(strings.ToLower(string(plain)))
		start := runeSliceIndex(lower, needle)
		if start >= 0 && x >= start-2 && x < len(plain) {
			return true
		}
	}
	return false
}

func toolRowContainsAt(lines []string, index, x int, label string, width int) bool {
	for position, line := range lines {
		plain := ansi.Strip(line)
		if !toolAppLineContains(plain, label) {
			continue
		}
		if index >= position-2 && index <= position+3 && x >= 0 && x < width {
			return true
		}
	}
	return false
}

func toolAppLineContains(line, label string) bool {
	for _, field := range strings.Fields(strings.ToLower(line)) {
		field = strings.Trim(field, "│")
		if field == "" {
			continue
		}
		return field == strings.ToLower(label)
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
