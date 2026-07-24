package tui

// selector is the compact non-filtering selector used by the mobile screens.
// Keeping it local avoids bubbles/list and its textinput/clipboard init while
// retaining Bubble Tea for the event loop and other Bubbles components.
type selector struct {
	index  int
	count  int
	width  int
	height int
}

func (s selector) Index() int { return s.index }

func (s *selector) Select(index int) {
	if s.count == 0 {
		s.index = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= s.count {
		index = s.count - 1
	}
	s.index = index
}

func (s *selector) CursorDown() { s.Select(s.index + 1) }
func (s *selector) CursorUp()   { s.Select(s.index - 1) }

func (s *selector) SetSize(width, height int) {
	s.width = max(1, width)
	s.height = max(1, height)
}
