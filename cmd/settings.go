package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thesubh213/winitrix/pkg/config"
)

var (
	configPath      string
	profileName     string
	selectManagers  bool
	notifyEnabled   bool
	notifyOnFailure bool
	retryBackoff    bool
	maxRetryDelay   int

	activeSettings config.EffectiveSettings
	configFound    bool
)

func loadSettings(cmd *cobra.Command, _ []string) error {
	if shouldSkipConfig(cmd) {
		activeSettings = config.Effective(config.Config{}, profileName)
		activeSettings.ConfigPath = configPath
		activeSettings.Profile = profileName
		return nil
	}

	cfg, resolvedPath, found, err := config.Load(configPath)
	if err != nil {
		// A corrupt config must not make every command (including --version
		// and doctor) unusable: warn and continue with built-in defaults.
		fmt.Fprintf(os.Stderr, "winitrix: warning: could not load config %s: %v\n", configPath, err)
		cfg = config.Config{}
		resolvedPath = configPath
		found = false
	}

	configFound = found

	profile := config.ResolveProfile(cfg, profileName)
	if profileName != "" {
		if _, ok := cfg.Profiles[profileName]; !ok {
			fmt.Fprintf(os.Stderr, "winitrix: warning: profile %q not found; using defaults\n", profileName)
		}
	}
	settings := config.Effective(cfg, profile)
	settings.Profile = profile
	settings.ConfigPath = resolvedPath

	applyFlagOverrides(cmd, &settings)
	activeSettings = settings

	return nil
}

func shouldSkipConfig(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	if cmd.Name() == "config" {
		return true
	}
	if parent := cmd.Parent(); parent != nil && parent.Name() == "config" {
		return true
	}
	return false
}

func applyFlagOverrides(cmd *cobra.Command, settings *config.EffectiveSettings) {
	if cmd == nil || settings == nil {
		return
	}

	overrideBool := func(name string, target *bool, value bool) {
		if cmd.Flags().Changed(name) {
			*target = value
		}
	}
	overrideInt := func(name string, target *int, value int) {
		if cmd.Flags().Changed(name) {
			*target = value
		}
	}
	overrideSlice := func(name string, target *[]string, value []string) {
		if cmd.Flags().Changed(name) {
			*target = append([]string(nil), value...)
		}
	}

	overrideBool("dry-run", &settings.DryRun, dryRun)
	overrideBool("nonstop", &settings.NonStop, nonstop)
	overrideInt("timeout", &settings.TimeoutMinutes, timeout)
	overrideInt("retries", &settings.Retries, retries)
	overrideInt("retry-delay", &settings.RetryDelaySeconds, retryDelay)
	overrideBool("retry-backoff", &settings.RetryBackoff, retryBackoff)
	overrideInt("max-retry-delay", &settings.MaxRetryDelaySeconds, maxRetryDelay)
	overrideSlice("only", &settings.Only, onlyManagers)
	overrideSlice("exclude", &settings.Exclude, excludeManagers)
	overrideBool("silent", &settings.Silent, silent)
	overrideBool("json", &settings.JSON, jsonOut)
	overrideBool("verbose", &settings.Verbose, verbose)
	overrideBool("select", &settings.SelectManagers, selectManagers)
	overrideBool("notify", &settings.NotificationsEnabled, notifyEnabled)
	overrideBool("notify-on-failure", &settings.NotifyOnlyOnFailure, notifyOnFailure)

	if cmd.Flags().Changed("profile") {
		settings.Profile = profileName
	}
}

func ActiveSettings() config.EffectiveSettings {
	return activeSettings
}

func ConfigSummary() string {
	if !configFound {
		return "default"
	}
	parts := []string{"config"}
	if activeSettings.Profile != "" {
		parts = append(parts, "profile="+activeSettings.Profile)
	}
	return fmt.Sprintf("%s (%s)", activeSettings.ConfigPath, strings.Join(parts, ", "))
}
