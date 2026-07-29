package tui

import "github.com/ericklucioh/mobdesk/internal/i18n"

func (m Model) renderTools() string {
	if !m.canManageHost() {
		return renderPage(m.text(i18n.TUIToolsTag, nil), m.text(i18n.TUIHostOnlyTitle, nil), m.text(i18n.TUIHostOnlyTools, nil))
	}
	width := contentWidth(m.width)
	items := toolListItemsLocalized(m.status, m.installingTool, m.localizer)
	m.toolsList.SetSize(width, max(6, m.height-8))
	view := renderToolItemsLocalized(items, m.toolsList.Index(), width, max(6, m.height-8), m.localizer)
	return tagStyle.Render(m.text(i18n.TUIToolsTag, nil)) + "\n" + titleStyle.Render(m.text(i18n.TUIToolsTitle, nil)) + "\n" + wrapText(m.text(i18n.TUIToolsBody, nil), width) + "\n\n" + view
}

// renderToolsBubbles preserva o nome usado por integrações e testes antigos.
func (m Model) renderToolsBubbles() string { return m.renderTools() }
