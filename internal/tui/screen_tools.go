package tui

func (m Model) renderTools() string {
	width := contentWidth(m.width)
	items := toolListItems(m.status, m.installingTool)
	m.toolsList.count = len(items)
	m.toolsList.SetSize(width, max(6, m.height-8))
	view := renderToolItems(items, m.toolsList.Index(), width, max(6, m.height-8))
	return tagStyle.Render("FERRAMENTAS UBUNTU") + "\n" + titleStyle.Render("Apps e linguagens") + "\n" + wrapText("Toque em uma linha para instalar · Enter confirma", width) + "\n\n" + view
}

// renderToolsBubbles preserva o nome usado por integrações e testes antigos.
func (m Model) renderToolsBubbles() string { return m.renderTools() }
