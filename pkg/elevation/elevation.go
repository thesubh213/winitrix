//go:build windows

package elevation

import (
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

// IsAdmin checks if the current process is running with administrative privileges.
func IsAdmin() bool {
	token := windows.GetCurrentProcessToken()
	return token.IsElevated()
}

// RunElevated relaunches the current executable with administrator privileges.
func RunElevated(args ...string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	verb := "runas"
	cwd, _ := os.Getwd()

	// Convert strings to UTF16 for the Windows API
	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)

	// Combine args
	var argString string
	for _, a := range args {
		argString += " " + a
	}
	argPtr, _ := syscall.UTF16PtrFromString(argString)

	err = windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, windows.SW_NORMAL)
	if err != nil {
		return err
	}

	// Exit the current non-elevated process so the elevated one takes over.
	os.Exit(0)
	return nil // unreachable, satisfies compiler
}

// CheckAdminRequirement checks output for common access-denied strings.
func CheckAdminRequirement(output string) bool {
	lower := strings.ToLower(output)
	patterns := []string{"0x80070005", "access is denied", "administrator privileges required"}
	return containsAny(lower, patterns)
}
