package tui

import (
	"os"
	"strings"
)

func (m Model) renderLogs() string {
	var builder strings.Builder
	builder.WriteString("HISTÓRICO\n\nLogs recentes\n\n" + m.focusAction(0, "[R] atualizar logs") + "\n" + m.focusAction(1, "[Esc] voltar") + "\n\n")
	if len(m.status.Installations) == 0 {
		return builder.String() + "Nenhuma instalação registrada.\n\nOs logs são mantidos em ~/.local/share/mobdesk/logs."
	}
	for _, record := range m.status.Installations {
		builder.WriteString(record.Name + "  " + record.State + "\n" + record.LogPath + "\n")
		if content, err := os.ReadFile(record.LogPath); err == nil {
			text := strings.TrimSpace(string(content))
			if len(text) > 300 {
				text = text[len(text)-300:]
			}
			builder.WriteString(text + "\n")
		}
		builder.WriteString("\n")
	}
	return builder.String()
}
