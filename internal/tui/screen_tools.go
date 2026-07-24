package tui

import (
	"charm.land/bubbles/v2/table"
	"github.com/ericklucioh/mobdesk/internal/install"
)

func (m Model) renderTools() string {
	width := contentWidth(m.width)
	// As larguras incluem somente o conteúdo; a tabela acrescenta padding
	// interno. Reservamos quatro colunas para esse padding no modo compacto.
	columns := []table.Column{{Title: "App", Width: max(8, width-12)}, {Title: "Estado", Width: 8}}
	if width >= 60 {
		columns = []table.Column{{Title: "App", Width: 16}, {Title: "Pacote", Width: 20}, {Title: "Estado", Width: 12}}
	}
	if width >= 72 {
		columns = []table.Column{{Title: "App", Width: 15}, {Title: "Pacote", Width: 18}, {Title: "Executável", Width: 15}, {Title: "Estado", Width: 12}}
	}
	// A tabela v2 recalcula o viewport imediatamente ao trocar colunas;
	// limpe as linhas antes para permitir o layout responsivo sem mismatch.
	m.toolsTable.SetRows(nil)
	m.toolsTable.SetColumns(columns)
	m.toolsTable.SetRows(toolRows(m.statusLoaded, width))
	m.toolsTable.SetWidth(width)
	m.toolsTable.SetHeight(max(6, m.height-8))
	view := m.toolsTable.View()
	if m.selectedTool >= 0 && m.selectedTool < len(install.Languages()) {
		item := install.Languages()[m.selectedTool]
		view += "\n" + mutedStyle.Render(item.Package+" · "+item.Executable)
	}
	return tagStyle.Render("FERRAMENTAS UBUNTU") + "\n" + titleStyle.Render("Apps e linguagens") + "\n" + wrapText("Toque em uma linha para instalar · Enter confirma", width) + "\n\n" + view
}

// renderToolsBubbles preserva o nome usado por integrações e testes antigos.
func (m Model) renderToolsBubbles() string { return m.renderTools() }
