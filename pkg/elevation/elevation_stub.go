//go:build !windows

package elevation

import "errors"

// IsAdmin always returns false on non-Windows platforms.
func IsAdmin() bool {
	return false
}

// RunElevated is not supported on non-Windows platforms.
func RunElevated(args ...string) error {
	return errors.New("elevation not supported on this platform")
}

// CheckAdminRequirement checks output for common access-denied strings.
func CheckAdminRequirement(output string) bool {
	return false
}
