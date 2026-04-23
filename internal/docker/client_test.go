package docker

import "testing"

func TestExtractDockerError(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "empty output",
			output: "",
			want:   "",
		},
		{
			name:   "only whitespace",
			output: "   \n\n  \n",
			want:   "",
		},
		{
			name: "network not found error (real case)",
			output: ` Network avita-network  Creating
 Network avita-network  Error
network avita-network declared as external, but could not be found`,
			want: "network avita-network declared as external, but could not be found",
		},
		{
			name: "error response from daemon is cleaned",
			output: `Error response from daemon: pull access denied for private/image`,
			want:   "pull access denied for private/image",
		},
		{
			name: "ERROR prefix is stripped",
			output: `ERROR: failed to solve: image not found`,
			want:   "failed to solve: image not found",
		},
		{
			name: "prioritizes error line over progress",
			output: ` Container foo Created
 Container foo Error: something broke
 Container foo Stopped`,
			want: "Container foo Error: something broke",
		},
		{
			name: "falls back to last non-progress line when no error keyword",
			output: ` Container foo Created
[+] Running 1/1
 Network bar Started
trailing informational line`,
			want: "trailing informational line",
		},
		{
			name: "strips ANSI escape codes",
			output: "\x1b[31mError: permission denied\x1b[0m",
			want:   "permission denied",
		},
		{
			name:   "returns empty when every line is benign progress",
			output: " Container foo Created\n Network bar Started\n[+] Running",
			want:   "",
		},
		{
			name:   "case-insensitive match on FAILED",
			output: "Build FAILED because of missing deps",
			want:   "Build FAILED because of missing deps",
		},
		{
			name:   "matches cannot keyword",
			output: "cannot connect to docker daemon",
			want:   "cannot connect to docker daemon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDockerError(tt.output)
			if got != tt.want {
				t.Errorf("extractDockerError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCleanDockerLine(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Error response from daemon: xyz", "xyz"},
		{"Error: foo", "foo"},
		{"ERROR: bar", "bar"},
		{"regular line", "regular line"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := cleanDockerLine(tt.in); got != tt.want {
				t.Errorf("cleanDockerLine(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsBenignProgressLine(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{" Container foo Created", true},
		{" Network bar Started", true},
		{" Volume vol Running", true},
		{" Image img Pulled", true},
		{" Image img Pulling", true},
		{"[+] Running 3/3", true},
		{"some actual error", false},
		{"Container foo Error", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := isBenignProgressLine(tt.in); got != tt.want {
				t.Errorf("isBenignProgressLine(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseConflictContainerName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "real compose error",
			in:   `Conflict. The container name "/sgo-backoffice-api" is already in use by container "0cbdac57a238"`,
			want: "sgo-backoffice-api",
		},
		{
			name: "without leading slash",
			in:   `container name "foo-bar" is already in use by container "abc"`,
			want: "foo-bar",
		},
		{
			name: "wrapped in daemon prefix",
			in:   `Error response from daemon: Conflict. The container name "/x" is already in use by container "y"`,
			want: "x",
		},
		{
			name: "not a conflict error",
			in:   "no such container",
			want: "",
		},
		{
			name: "empty",
			in:   "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseConflictContainerName(tt.in); got != tt.want {
				t.Errorf("parseConflictContainerName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestUnavailableStatus(t *testing.T) {
	s := UnavailableStatus()
	if s.Symbol != "!" {
		t.Errorf("Symbol = %q, want %q", s.Symbol, "!")
	}
	if s.Label != "docker indisponivel" {
		t.Errorf("Label = %q, want %q", s.Label, "docker indisponivel")
	}
	if s.Total != 0 || s.Running != 0 {
		t.Errorf("expected zeroed counters, got total=%d running=%d", s.Total, s.Running)
	}
}
