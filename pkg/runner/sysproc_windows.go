//go:build windows

package runner

import (
	"os/exec"
	"syscall"
)

func setSysProcAttr(cmd *exec.Cmd) {
	// Hide window to prevent console flashing on Windows.
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
