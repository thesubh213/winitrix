package ecosystem

import "strings"

var managerAliases = map[string][]string{
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

	var filtered []PackageManager
	for _, mgr := range managers {
		name := mgr.Name()
		if len(only) > 0 && !matchesAnyToken(name, only) {
			continue
		}
		if len(exclude) > 0 && matchesAnyToken(name, exclude) {
			continue
		}
		filtered = append(filtered, mgr)
	}

	return filtered
}

func matchesAnyToken(managerName string, tokens []string) bool {
	for _, token := range tokens {
		if managerMatchesToken(managerName, token) {
			return true
		}
	}
	return false
}

func managerMatchesToken(managerName, token string) bool {
	nameNorm := normalizeToken(managerName)
	tokenNorm := normalizeToken(token)
	if tokenNorm == "" {
		return false
	}

	if strings.Contains(nameNorm, tokenNorm) {
		return true
	}

	if aliases, ok := managerAliases[tokenNorm]; ok {
		for _, alias := range aliases {
			if strings.Contains(nameNorm, alias) {
				return true
			}
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
