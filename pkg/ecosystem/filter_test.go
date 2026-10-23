package ecosystem

import (
	"testing"
)

func filterNames(managers []PackageManager, only, exclude []string) []string {
	filtered := FilterManagers(managers, only, exclude)
	names := make([]string, 0, len(filtered))
	for _, m := range filtered {
		names = append(names, m.Name())
	}
	return names
}

func TestFilterManagersExactMatchWins(t *testing.T) {
	managers := []PackageManager{
		NewWingetManager(),
		NewScoopManager(),
		NewNpmManager(),
		NewPipManager(),
		NewPipxManager(),
		NewGoManager(),
		NewDotnetManager(),
	}

	tests := []struct {
		name     string
		only     []string
		exclude  []string
		expected []string
	}{
		{
			name:     "pip selects only Pip, not pipx",
			only:     []string{"pip"},
			expected: []string{"Pip"},
		},
		{
			name:     "pipx selects only pipx",
			only:     []string{"pipx"},
			expected: []string{"pipx"},
		},
		{
			name:     "case insensitive token",
			only:     []string{"PIP"},
			expected: []string{"Pip"},
		},
		{
			name:     "dotnet alias resolves to .NET Tools",
			only:     []string{"dotnet"},
			expected: []string{".NET Tools"},
		},
		{
			name:     "normalized name nettools matches .NET Tools",
			only:     []string{"net tools"},
			expected: []string{".NET Tools"},
		},
		{
			name:     "go alias resolves to Go Tools only",
			only:     []string{"go"},
			expected: []string{"Go Tools"},
		},
		{
			name:     "substring fallback still works for partial tokens",
			only:     []string{"wing"},
			expected: []string{"Winget"},
		},
		{
			name:     "exclude removes only the exact manager",
			exclude:  []string{"pip"},
			expected: []string{"Winget", "Scoop", "Npm", "pipx", "Go Tools", ".NET Tools"},
		},
		{
			name:     "whitespace-only tokens are treated as no filter",
			only:     []string{"  "},
			expected: []string{"Winget", "Scoop", "Npm", "Pip", "pipx", "Go Tools", ".NET Tools"},
		},
		{
			name:     "multiple tokens combine",
			only:     []string{"npm", "scoop"},
			expected: []string{"Scoop", "Npm"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterNames(managers, tt.only, tt.exclude)
			if len(got) != len(tt.expected) {
				t.Fatalf("got %v, want %v", got, tt.expected)
			}
			gotSet := make(map[string]bool, len(got))
			for _, n := range got {
				gotSet[n] = true
			}
			for _, want := range tt.expected {
				if !gotSet[want] {
					t.Fatalf("got %v, want it to contain %q", got, want)
				}
			}
		})
	}
}

func TestFilterManagersNoFiltersReturnsAll(t *testing.T) {
	managers := []PackageManager{NewWingetManager(), NewNpmManager()}
	if got := filterNames(managers, nil, nil); len(got) != 2 {
		t.Fatalf("expected all managers, got %v", got)
	}
}
