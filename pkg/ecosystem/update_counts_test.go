package ecosystem

import "testing"

func TestCountVersionChanges(t *testing.T) {
	before := map[string]string{
		"black":  "24.2.0",
		"httpie": "3.2.1",
		"yarn":   "1.22.22",
	}
	after := map[string]string{
		"black":  "24.3.0",
		"httpie": "3.2.1",
		"yarn":   "1.22.22",
		"extra":  "1.0.0",
	}

	if got := countVersionChanges(before, after); got != 1 {
		t.Fatalf("countVersionChanges() = %d, want 1", got)
	}
}

func TestParsePipxVersions(t *testing.T) {
	output := `{
  "venvs": {
    "black": {
      "metadata": {
        "main_package": {
          "package": "black",
          "package_version": "24.3.0"
        }
      }
    },
    "httpie": {
      "metadata": {
        "main_package": {
          "package": "httpie",
          "package_version": "3.2.2"
        }
      }
    }
  }
}`

	versions, ok := parsePipxVersions(output)
	if !ok {
		t.Fatal("parsePipxVersions() reported failure")
	}

	if got := versions["black"]; got != "24.3.0" {
		t.Fatalf("black version = %q, want %q", got, "24.3.0")
	}

	if got := versions["httpie"]; got != "3.2.2" {
		t.Fatalf("httpie version = %q, want %q", got, "3.2.2")
	}
}

func TestParseYarnGlobalVersions(t *testing.T) {
	output := `yarn global v1.22.22
info No lockfile found.
├─ @scope/tool@1.2.3
├─ nodemon@3.1.10
└─ typescript@5.4.5
Done in 0.42s.`

	versions, ok := parseYarnGlobalVersions(output)
	if !ok {
		t.Fatal("parseYarnGlobalVersions() reported failure")
	}

	if got := versions["@scope/tool"]; got != "1.2.3" {
		t.Fatalf("@scope/tool version = %q, want %q", got, "1.2.3")
	}

	if got := versions["nodemon"]; got != "3.1.10" {
		t.Fatalf("nodemon version = %q, want %q", got, "3.1.10")
	}

	if got := versions["typescript"]; got != "5.4.5" {
		t.Fatalf("typescript version = %q, want %q", got, "5.4.5")
	}
}
