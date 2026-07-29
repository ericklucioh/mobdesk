package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

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
func blockContainsAt(lines []string, index, x int, text string) bool {
	for position, line := range lines {
		plain := ansi.Strip(line)
		start := strings.Index(strings.ToLower(plain), strings.ToLower(text))
		// Botões Lipgloss têm três linhas: borda superior, conteúdo e borda
		// inferior. Todas pertencem ao mesmo alvo.
		if start < 0 || index < position-1 || index > position+1 {
			continue
		}
		first := utf8.RuneCountInString(plain[:start])
		last := first + utf8.RuneCountInString(text) - 1
		// O alvo inclui o padding e a borda do botão, não apenas as letras.
		return x >= first-2 && x <= last+2
	}
	return false
}

// blockContainsAtAny é usado quando uma tela repete um rótulo, como "Status"
// no cartão da Home e no resumo da workstation. Ele testa todas as ocorrências
// antes de desistir, mantendo a validação horizontal do alvo.
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

func toolRowContainsAt(lines []string, index, x int, label string, width int) bool {
	for position, line := range lines {
		plain := ansi.Strip(line)
		if !containsToolLabel(plain, label) {
			continue
		}
		if index >= position && index <= position+1 && x >= 0 && x < width {
			return true
		}
	}
	return false
}

func containsToolLabel(line, label string) bool {
	line = strings.ToLower(line)
	label = strings.ToLower(label)
	start := 0
	for start < len(line) {
		match := strings.Index(line[start:], label)
		if match < 0 {
			return false
		}
		match += start
		end := match + len(label)
		if (match == 0 || !toolLabelChar(line[match-1])) && (end == len(line) || !toolLabelChar(line[end])) {
			return true
		}
		start = end
	}
	return false
}

func toolLabelChar(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= '0' && value <= '9' || value == '_'
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
