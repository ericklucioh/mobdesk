package tui

import tea "charm.land/bubbletea/v2"

func (m Model) updateMouseDrag(mouse tea.Mouse) (tea.Model, tea.Cmd) {
	if !m.dragging || mouse.Y == m.dragY {
		return m, nil
	}
	delta := m.dragY - mouse.Y
	m.dragY = mouse.Y
	if m.screen == toolsScreen {
		return m.dragTable(delta)
	}
	if delta > 0 {
		m.viewport.ScrollDown(delta)
	} else {
		m.viewport.ScrollUp(-delta)
	}
	return m, nil
}

func (m Model) dragTable(delta int) (tea.Model, tea.Cmd) {
	button := tea.MouseWheelUp
	if delta > 0 {
		button = tea.MouseWheelDown
	}
	for step := 0; step < abs(delta); step++ {
		var cmd tea.Cmd
		m.toolsTable, cmd = m.toolsTable.Update(tea.MouseWheelMsg{Button: button})
		if cmd != nil {
			return m, cmd
		}
	}
	return m, nil
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
