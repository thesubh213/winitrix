package doctor

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestCheckPathVariables_ContainsEssentialPaths(t *testing.T) {
	// Save and restore PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set a PATH that contains all essential entries
	os.Setenv("PATH", `C:\Windows\system32;C:\Windows;C:\Windows\System32\Wbem;C:\Windows\System32\WindowsPowerShell\v1.0`)

	ok, msg := CheckPathVariables(context.Background())
	if !ok {
		t.Errorf("Expected healthy PATH, got: %s", msg)
	}
	if !strings.Contains(msg, "healthy") {
		t.Errorf("Expected 'healthy' in message, got: %s", msg)
	}
}

func TestCheckPathVariables_MissingEntries(t *testing.T) {
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set a PATH missing some essential entries
	os.Setenv("PATH", `C:\SomeOtherDir`)

	ok, msg := CheckPathVariables(context.Background())
	if ok {
		t.Errorf("Expected unhealthy PATH with missing entries, but got healthy")
	}
	if !strings.Contains(msg, "Missing essential PATH entries") {
		t.Errorf("Expected missing entries message, got: %s", msg)
	}
}
