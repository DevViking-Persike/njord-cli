package components

import tea "github.com/charmbracelet/bubbletea"

// ListNav handles up/down navigation. Returns new cursor value.
func ListNav(msg tea.KeyMsg, cursor, total int) (int, bool) {
	switch msg.String() {
	case "up", "k":
		if cursor > 0 {
			return cursor - 1, true
		}
	case "down", "j":
		if cursor < total-1 {
			return cursor + 1, true
		}
	}
	return cursor, false
}

// TextInput handles standard text input (backspace, ctrl+u, runes).
// filter is optional (e.g., DigitsOnly). Returns new buffer.
func TextInput(msg tea.KeyMsg, buf string, filter func(rune) bool) (string, bool) {
	switch msg.String() {
	case "backspace":
		if len(buf) > 0 {
			return buf[:len(buf)-1], true
		}
	case "ctrl+u":
		return "", true
	default:
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
			var added string
			for _, r := range msg.Runes {
				if filter == nil || filter(r) {
					added += string(r)
				}
			}
			if added != "" {
				return buf + added, true
			}
		}
	}
	return buf, false
}

// DigitsOnly is a filter that only allows digit characters.
func DigitsOnly(r rune) bool { return r >= '0' && r <= '9' }
