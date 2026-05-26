package cmd

import (
	"testing"

	"github.com/thesubh213/winitrix/pkg/config"
)

func TestBuildConfigShowOutput(t *testing.T) {
	settings := config.EffectiveSettings{
		ConfigPath:           `C:\Users\subhr\AppData\Roaming\Winitrix\config.toml`,
		Profile:              "default",
		DryRun:               true,
		Silent:               false,
		JSON:                 false,
		Verbose:              true,
		NonStop:              true,
		TimeoutMinutes:       15,
		Retries:              2,
		RetryDelaySeconds:    4,
		RetryBackoff:         true,
		MaxRetryDelaySeconds: 45,
		SelectManagers:       true,
		NotificationsEnabled: true,
		NotifyOnlyOnFailure:  false,
		Only:                 []string{"winget", "scoop"},
		Exclude:              []string{"wsl"},
	}

	output := buildConfigShowOutput(settings, true)
	if !output.Loaded {
		t.Fatal("expected loaded output")
	}
	if output.Source == "" {
		t.Fatal("expected non-empty source")
	}
	if output.Settings.Profile != "default" {
		t.Fatalf("profile = %q, want %q", output.Settings.Profile, "default")
	}
	if len(output.Settings.Only) != 2 {
		t.Fatalf("only managers length = %d, want 2", len(output.Settings.Only))
	}
}
