package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const appName = "Winitrix"

// ConfigDir returns the directory for user-scoped configuration.
func ConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err == nil && dir != "" {
		return filepath.Join(dir, appName), nil
	}

	if fallback := os.Getenv("APPDATA"); fallback != "" {
		return filepath.Join(fallback, appName), nil
	}
	if fallback := os.Getenv("USERPROFILE"); fallback != "" {
		return filepath.Join(fallback, "AppData", "Roaming", appName), nil
	}

	return "", fmt.Errorf("unable to determine config directory")
}

// CacheDir returns the directory for cache and state files.
func CacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err == nil && dir != "" {
		return filepath.Join(dir, appName), nil
	}

	if fallback := os.Getenv("LOCALAPPDATA"); fallback != "" {
		return filepath.Join(fallback, appName), nil
	}
	if fallback := os.Getenv("USERPROFILE"); fallback != "" {
		return filepath.Join(fallback, "AppData", "Local", appName), nil
	}

	return "", fmt.Errorf("unable to determine cache directory")
}

// EnsureDir creates the directory if missing.
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// DefaultStatePath returns the default state file path.
func DefaultStatePath() (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}

// DefaultReportPath returns the default last report file path.
func DefaultReportPath() (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "last_report.json"), nil
}
