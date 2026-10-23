package ecosystem

import "strings"

// managerAliasTokens maps a user-facing token to normalized manager names it
// should select even though the names don't share a prefix. Exact matching is
// preferred over substring matching so `--only pip` never selects pipx.
var managerAliasTokens = map[string][]string{
	"dotnet": {"nettools"},
	"vscode": {"vscodeextensions"},
	"wsl":    {"wslapt"},
	"go":     {"gotools"},
}

// FilterManagers filters managers by optional allow/deny lists.
func FilterManagers(managers []PackageManager, only []string, exclude []string) []PackageManager {
	if len(only) == 0 && len(exclude) == 0 {
		return managers
	}

	onlyMatchers := buildMatchers(managers, only)
	excludeMatchers := buildMatchers(managers, exclude)

	var filtered []PackageManager
	for _, mgr := range managers {
		if len(onlyMatchers) > 0 && !matchesAny(onlyMatchers, mgr.Name()) {
			continue
		}
		if len(excludeMatchers) > 0 && matchesAny(excludeMatchers, mgr.Name()) {
			continue
		}
		filtered = append(filtered, mgr)
	}

	return filtered
}

// tokenMatcher matches normalized manager names. When a token resolves to at
// least one manager by exact name (or alias), only those managers match;
// otherwise the token falls back to substring matching for convenience.
type tokenMatcher struct {
	exactNames map[string]bool
	substring  string
}

func buildMatchers(managers []PackageManager, tokens []string) []tokenMatcher {
	matchers := make([]tokenMatcher, 0, len(tokens))
	for _, token := range tokens {
		norm := normalizeToken(token)
		if norm == "" {
			continue
		}

		exact := make(map[string]bool)
		for _, mgr := range managers {
			name := normalizeToken(mgr.Name())
			if name == norm || aliasMatches(norm, name) {
				exact[name] = true
			}
		}

		if len(exact) > 0 {
			matchers = append(matchers, tokenMatcher{exactNames: exact})
		} else {
			matchers = append(matchers, tokenMatcher{substring: norm})
		}
	}
	return matchers
}

func aliasMatches(tokenNorm, managerNorm string) bool {
	for _, alias := range managerAliasTokens[tokenNorm] {
		if alias == managerNorm {
			return true
		}
	}
	return false
}

func matchesAny(matchers []tokenMatcher, managerName string) bool {
	name := normalizeToken(managerName)
	for _, m := range matchers {
		if m.exactNames != nil {
			if m.exactNames[name] {
				return true
			}
			continue
		}
		if m.substring != "" && strings.Contains(name, m.substring) {
			return true
		}
	}
	return false
}

func normalizeToken(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
