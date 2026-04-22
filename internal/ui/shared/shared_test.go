package shared

import (
	"strings"
	"testing"
	"time"
)

func TestNjordTitle_ContainsRuneAndName(t *testing.T) {
	out := NjordTitle()
	if !strings.Contains(out, "ᚾ") {
		t.Errorf("expected rune ᚾ in output, got %q", out)
	}
	if !strings.Contains(out, "N J O R D") {
		t.Errorf("expected name 'N J O R D' in output, got %q", out)
	}
}

func TestTimeAgo(t *testing.T) {
	now := time.Now()
	tests := []struct {
		t    time.Time
		want string
	}{
		{now.Add(-30 * time.Second), "agora"},
		{now.Add(-5 * time.Minute), "5m atrás"},
		{now.Add(-3 * time.Hour), "3h atrás"},
		{now.Add(-25 * time.Hour), "ontem"},
		{now.Add(-72 * time.Hour), "3d atrás"},
	}
	for _, tt := range tests {
		if got := TimeAgo(tt.t); got != tt.want {
			t.Errorf("TimeAgo(%v) = %q, want %q", tt.t, got, tt.want)
		}
	}
}

func TestConstants(t *testing.T) {
	if MinCardWidth <= 0 || BorderOverhead < 0 || CardHeight <= 0 {
		t.Errorf("unexpected constants: min=%d border=%d height=%d", MinCardWidth, BorderOverhead, CardHeight)
	}
}
