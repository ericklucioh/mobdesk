package tui

import "fmt"

func (m Model) renderSystem() string {
	return renderPage("SISTEMA", "Mobdesk "+m.version.Version, fmt.Sprintf("Canal: %s\nPlataforma: %s/%s\n\n%s\n%s", m.version.Channel, m.version.OS, m.version.Architecture, m.focusAction(0, "[V] verificar atualização"), m.focusAction(1, "[A] aplicar atualização")))
}
