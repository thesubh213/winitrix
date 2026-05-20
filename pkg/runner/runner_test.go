package runner

import "testing"

func TestTruncateOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"Short string unchanged", "hello", 10, "hello"},
		{"Exact length unchanged", "hello", 5, "hello"},
		{"Long string truncated to tail", "abcdefghij", 5, "fghij"},
		{"Empty string", "", 5, ""},
		{"Max 0", "hello", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateOutput(tt.input, tt.maxLen)
			if got != tt.expected {
				t.Errorf("truncateOutput(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.expected)
			}
		})
	}
}
