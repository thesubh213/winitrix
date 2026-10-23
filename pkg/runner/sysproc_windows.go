//go:build windows

package runner

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

const createNewProcessGroup = 0x00000200

func setSysProcAttr(cmd *exec.Cmd) {
	// Hide the window to prevent console flashing and isolate the child from
	// console Ctrl+C so cancellation stays under our control.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNewProcessGroup,
	}
}

// KillTree terminates a process and all of its descendants. Package managers
// wrap installers, git, or PowerShell children that would otherwise keep
// running after a timeout or user cancellation.
func KillTree(pid int) error {
	kill := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	kill.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := kill.Run(); err != nil {
		if proc, perr := os.FindProcess(pid); perr == nil {
			return proc.Kill()
		}
		return err
	}
	return nil
}
