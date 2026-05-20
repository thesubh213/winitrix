package doctor

import (
	"context"
	"os"
	"strings"
)

// CheckPathVariables verifies that essential directories are in the PATH.
func CheckPathVariables(ctx context.Context) (bool, string) {
	path := os.Getenv("PATH")
	var missing []string

	essentialPaths := []string{
		"Windows\\system32",
		"Windows",
		"Windows\\System32\\Wbem",
		"Windows\\System32\\WindowsPowerShell\\v1.0",
	}

	for _, p := range essentialPaths {
		if !strings.Contains(strings.ToLower(path), strings.ToLower(p)) {
			missing = append(missing, p)
		}
	}

	if len(missing) > 0 {
		return false, "Missing essential PATH entries: " + strings.Join(missing, ", ")
	}
	return true, "System PATH looks healthy."
}
