package components

// ScrollState manages viewport scrolling.
type ScrollState struct {
	Offset int
	Height int // total terminal height
}

// VisibleRows returns rows available for list items.
func (s ScrollState) VisibleRows(chromeLines int) int {
	available := s.Height - chromeLines
	if available < 3 {
		return 3
	}
	return available
}

// EnsureVisible adjusts offset to keep cursor in view.
func (s *ScrollState) EnsureVisible(cursor, chromeLines int) {
	visible := s.VisibleRows(chromeLines)
	if cursor < s.Offset {
		s.Offset = cursor
	}
	if cursor >= s.Offset+visible {
		s.Offset = cursor - visible + 1
	}
}

// Bounds returns (start, end) indices for visible slice of items.
func (s ScrollState) Bounds(total, chromeLines int) (int, int) {
	visible := s.VisibleRows(chromeLines)
	start := s.Offset
	end := start + visible
	if end > total {
		end = total
	}
	return start, end
}
