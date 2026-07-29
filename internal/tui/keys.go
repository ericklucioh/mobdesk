package tui

import "charm.land/bubbles/v2/key"
import "github.com/ericklucioh/mobdesk/internal/i18n"

type tuiHelpKeyMap struct{ localizer i18n.Localizer }

func (m tuiHelpKeyMap) text(id i18n.MessageID) string { return m.localizer.Text(id, nil) }

func (m tuiHelpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", m.text(i18n.TUIFooterFocus))),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", m.text(i18n.TUIFooterAct))),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", m.text(i18n.TUIFooterQuit))),
	}
}

func (tuiHelpKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{tuiHelpKeyMap{}.ShortHelp()} }
