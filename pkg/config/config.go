package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/thesubh213/winitrix/pkg/paths"
)

const currentVersion = 1

// Config is the root configuration structure.
type Config struct {
	Version        int                `toml:"version"`
	DefaultProfile string             `toml:"default_profile,omitempty"`
	Defaults       Defaults           `toml:"defaults"`
	Profiles       map[string]Profile `toml:"profiles,omitempty"`
	UI             UIConfig           `toml:"ui"`
	Notifications  NotificationConfig `toml:"notifications"`
}

// Defaults defines the update behavior defaults.
type Defaults struct {
	DryRun               *bool    `toml:"dry_run,omitempty"`
	NonStop              *bool    `toml:"nonstop,omitempty"`
	TimeoutMinutes       *int     `toml:"timeout_minutes,omitempty"`
	Retries              *int     `toml:"retries,omitempty"`
	RetryDelaySeconds    *int     `toml:"retry_delay_seconds,omitempty"`
	RetryBackoff         *bool    `toml:"retry_backoff,omitempty"`
	MaxRetryDelaySeconds *int     `toml:"max_retry_delay_seconds,omitempty"`
	Only                 []string `toml:"only,omitempty"`
	Exclude              []string `toml:"exclude,omitempty"`
	Silent               *bool    `toml:"silent,omitempty"`
	JSON                 *bool    `toml:"json,omitempty"`
	Verbose              *bool    `toml:"verbose,omitempty"`
}

// Profile lets users override defaults for a named profile.
type Profile struct {
	Defaults
}

// UIConfig configures the TUI behavior.
type UIConfig struct {
	SelectManagers *bool `toml:"select_managers,omitempty"`
}

// NotificationConfig controls desktop notifications.
type NotificationConfig struct {
	Enabled       *bool `toml:"enabled,omitempty"`
	OnlyOnFailure *bool `toml:"only_on_failure,omitempty"`
}

// EffectiveSettings are the merged settings used at runtime.
type EffectiveSettings struct {
	DryRun               bool
	NonStop              bool
	TimeoutMinutes       int
	Retries              int
	RetryDelaySeconds    int
	RetryBackoff         bool
	MaxRetryDelaySeconds int
	Only                 []string
	Exclude              []string
	Silent               bool
	JSON                 bool
	Verbose              bool
	SelectManagers       bool
	NotificationsEnabled bool
	NotifyOnlyOnFailure  bool
	Profile              string
	ConfigPath           string
}

// Load reads config from the provided path, or the default path if empty.
func Load(path string) (Config, string, bool, error) {
	resolved := path
	if resolved == "" {
		var err error
		resolved, err = paths.DefaultConfigPath()
		if err != nil {
			return Config{}, "", false, err
		}
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, resolved, false, nil
		}
		return Config{}, resolved, false, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, resolved, false, err
	}

	if cfg.Version == 0 {
		cfg.Version = currentVersion
	}

	return cfg, resolved, true, nil
}

// ResolveProfile returns the effective profile name.
func ResolveProfile(cfg Config, requested string) string {
	if requested != "" {
		return requested
	}
	return cfg.DefaultProfile
}

// Effective merges defaults and profile overrides into runtime settings.
func Effective(cfg Config, profileName string) EffectiveSettings {
	settings := EffectiveSettings{
		TimeoutMinutes:       10,
		Retries:              0,
		RetryDelaySeconds:    2,
		RetryBackoff:         false,
		MaxRetryDelaySeconds: 30,
		SelectManagers:       false,
		NotificationsEnabled: false,
		NotifyOnlyOnFailure:  true,
		Profile:              profileName,
	}

	applyDefaults(&settings, cfg.Defaults)

	if profileName != "" {
		if profile, ok := cfg.Profiles[profileName]; ok {
			applyDefaults(&settings, profile.Defaults)
		}
	}

	if cfg.UI.SelectManagers != nil {
		settings.SelectManagers = *cfg.UI.SelectManagers
	}
	if cfg.Notifications.Enabled != nil {
		settings.NotificationsEnabled = *cfg.Notifications.Enabled
	}
	if cfg.Notifications.OnlyOnFailure != nil {
		settings.NotifyOnlyOnFailure = *cfg.Notifications.OnlyOnFailure
	}

	return settings
}

func applyDefaults(target *EffectiveSettings, defaults Defaults) {
	if defaults.DryRun != nil {
		target.DryRun = *defaults.DryRun
	}
	if defaults.NonStop != nil {
		target.NonStop = *defaults.NonStop
	}
	if defaults.TimeoutMinutes != nil {
		target.TimeoutMinutes = *defaults.TimeoutMinutes
	}
	if defaults.Retries != nil {
		target.Retries = *defaults.Retries
	}
	if defaults.RetryDelaySeconds != nil {
		target.RetryDelaySeconds = *defaults.RetryDelaySeconds
	}
	if defaults.RetryBackoff != nil {
		target.RetryBackoff = *defaults.RetryBackoff
	}
	if defaults.MaxRetryDelaySeconds != nil {
		target.MaxRetryDelaySeconds = *defaults.MaxRetryDelaySeconds
	}
	if defaults.Only != nil {
		target.Only = append([]string(nil), defaults.Only...)
	}
	if defaults.Exclude != nil {
		target.Exclude = append([]string(nil), defaults.Exclude...)
	}
	if defaults.Silent != nil {
		target.Silent = *defaults.Silent
	}
	if defaults.JSON != nil {
		target.JSON = *defaults.JSON
	}
	if defaults.Verbose != nil {
		target.Verbose = *defaults.Verbose
	}
}

// WriteDefaultConfig writes a starter config to the provided path.
func WriteDefaultConfig(path string, force bool) (string, error) {
	resolved := path
	if resolved == "" {
		var err error
		resolved, err = paths.DefaultConfigPath()
		if err != nil {
			return "", err
		}
	}

	if !force {
		if _, err := os.Stat(resolved); err == nil {
			return "", os.ErrExist
		}
	}

	if err := paths.EnsureDir(filepath.Dir(resolved)); err != nil {
		return "", err
	}

	content := defaultConfigTemplate()
	return resolved, os.WriteFile(resolved, []byte(content), 0644)
}

func defaultConfigTemplate() string {
	return `# Winitrix configuration
version = 1

# default_profile = "default"

[defaults]
# dry_run = false
# nonstop = false
# silent = false
# json = false
# verbose = false
# timeout_minutes = 10
# retries = 0
# retry_delay_seconds = 2
# retry_backoff = false
# max_retry_delay_seconds = 30
# only = ["winget", "scoop"]
# exclude = ["wsl"]

[ui]
# select_managers = false

[notifications]
# enabled = false
# only_on_failure = true

[profiles.default]
# Example profile overrides.
# only = ["winget", "scoop"]
`
}
