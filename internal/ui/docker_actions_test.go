package ui

import "testing"

func TestTranslateDockerState(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"running", "rodando"},
		{"created", "criado"},
		{"exited", "parado"},
		{"paused", "pausado"},
		{"restarting", "reiniciando"},
		{"dead", "morto"},
		{"unknown-state", "unknown-state"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := translateDockerState(tt.in); got != tt.want {
				t.Errorf("translateDockerState(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
