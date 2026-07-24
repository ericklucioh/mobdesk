package tui

import "charm.land/bubbles/v2/key"

type tuiHelpKeyMap struct{}

func (tuiHelpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "foco")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "agir")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "sair")),
	}
}

func (tuiHelpKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{tuiHelpKeyMap{}.ShortHelp()} }
