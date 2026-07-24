package tui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
)

func renderPage(tag, title, body string) string {
	return tagStyle.Render(tag) + "\n" + titleStyle.Render(title) + "\n" + bodyStyle.Render(body)
}
func (m Model) focusAction(index int, text string) string {
	if m.focus == index {
		return "> " + text
	}
	return "  " + text
}
func available(value bool) string {
	if value {
		return "disponível"
	}
	return "inativo"
}
func yesNo(value bool) string {
	if value {
		return "sim"
	}
	return "não"
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
func blockContains(lines []string, index int, text string) bool {
	for position, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(text)) {
			return index >= max(0, position-2) && index < min(len(lines), position+3)
		}
	}
	return false
}

func blockContainsAt(lines []string, index, x int, text string) bool {
	for position, line := range lines {
		if !strings.Contains(strings.ToLower(line), strings.ToLower(text)) {
			continue
		}
		start := max(0, position-2)
		end := min(len(lines), position+3)
		if index < start || index >= end {
			return false
		}
		for row := start; row < end; row++ {
			plain := []rune(ansi.Strip(lines[row]))
			first, last := -1, -1
			for column, value := range plain {
				if !unicode.IsSpace(value) {
					if first == -1 {
						first = column
					}
					last = column
				}
			}
			if row == index && first >= 0 {
				return x >= first && x <= last
			}
		}
		return false
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
