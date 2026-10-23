package ecosystem

import "testing"

// TestExtractSubjectAfterKeywordExoticRune is a regression test: the keyword
// index was computed on a lowercased copy while the slice was taken from the
// original string. Lowercasing U+023A expands the byte length, which made the
// slice bounds invalid (panic) for crafted input.
func TestExtractSubjectAfterKeywordExoticRune(t *testing.T) {
	line := "Ⱥ installing"
	if got := extractSubjectAfterKeyword(line, "installing"); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestExtractSubjectAfterKeyword(t *testing.T) {
	tests := []struct {
		line    string
		keyword string
		want    string
	}{
		{"Fetching node-20.11.0", "fetching", "node-20.11.0"},
		{"Installing: pkg [1.0]", "installing", "pkg [1.0]"},
		{"  downloading  archive.tar.gz", "downloading", "archive.tar.gz"},
		{"no keyword here", "installing", ""},
		{"installing", "installing", ""},
	}

	for _, tt := range tests {
		if got := extractSubjectAfterKeyword(tt.line, tt.keyword); got != tt.want {
			t.Errorf("extractSubjectAfterKeyword(%q, %q) = %q, want %q", tt.line, tt.keyword, got, tt.want)
		}
	}
}

func TestParseProgressLineSamples(t *testing.T) {
	tests := []struct {
		name       string
		manager    string
		line       string
		wantStatus string
		wantPkg    string
		wantPct    int
	}{
		{
			name:       "winget installing",
			manager:    "Winget",
			line:       "Installing: 7zip [23.01]",
			wantStatus: "Installing",
			wantPkg:    "7zip",
			wantPct:    -1,
		},
		{
			name:       "winget download with percentage",
			manager:    "Winget",
			line:       "Downloading: aria2 [1.37.0 45%]",
			wantStatus: "Downloading",
			wantPkg:    "aria2",
			wantPct:    45,
		},
		{
			name:       "scoop update with version transition",
			manager:    "Scoop",
			line:       "Updating 'git' (2.44.0 -> 2.45.0)",
			wantStatus: "Updating",
			wantPkg:    "git",
			wantPct:    -1,
		},
		{
			name:       "npm version change",
			manager:    "Npm",
			line:       "typescript 5.3.0 -> 5.4.2",
			wantStatus: "Updating",
			wantPkg:    "typescript",
			wantPct:    -1,
		},
		{
			name:       "pip collecting",
			manager:    "pip",
			line:       "Collecting requests",
			wantStatus: "Collecting",
			wantPkg:    "requests",
			wantPct:    -1,
		},
		{
			name:       "chocolatey prepares upgrades",
			manager:    "Chocolatey",
			line:       "Upgrading the following packages",
			wantStatus: "Preparing package upgrades",
			wantPkg:    "",
			wantPct:    -1,
		},
		{
			name:       "cargo install",
			manager:    "cargo",
			line:       "Installing cargo-update",
			wantStatus: "Installing",
			wantPkg:    "cargo-update",
			wantPct:    -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update, ok := ParseProgressLine(tt.manager, tt.line)
			if !ok {
				t.Fatalf("ParseProgressLine(%q, %q) did not match", tt.manager, tt.line)
			}
			if update.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", update.Status, tt.wantStatus)
			}
			if update.PackageName != tt.wantPkg {
				t.Errorf("package = %q, want %q", update.PackageName, tt.wantPkg)
			}
			if tt.wantPct >= 0 && update.Percentage != tt.wantPct {
				t.Errorf("percentage = %d, want %d", update.Percentage, tt.wantPct)
			}
		})
	}
}

func TestParseProgressLineStripsAnsiAndCR(t *testing.T) {
	line := "\x1b[32mInstalling: vim [9.1]\x1b[0m\r"
	update, ok := ParseProgressLine("Winget", line)
	if !ok {
		t.Fatalf("ansi-wrapped line was not parsed")
	}
	if update.PackageName != "vim" {
		t.Fatalf("package = %q, want vim", update.PackageName)
	}
}

func TestParseProgressLineRejectsNoise(t *testing.T) {
	if _, ok := ParseProgressLine("Winget", "some random noise"); ok {
		t.Fatalf("noise line should not produce progress")
	}
	if _, ok := ParseProgressLine("Winget", ""); ok {
		t.Fatalf("empty line should not produce progress")
	}
}
