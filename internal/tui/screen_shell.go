package tui

func (m Model) renderShell() string {
	return renderPage("SHELL UBUNTU", "Abrir shell", "A TUI será suspensa enquanto o shell estiver aberto.\n\n"+m.focusAction(0, "[Enter] abrir shell")+"\n"+m.focusAction(1, "[Esc] voltar"))
}
