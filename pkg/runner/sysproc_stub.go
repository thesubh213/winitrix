//go:build !windows

package runner

import (
	"os"
	"os/exec"
)

func setSysProcAttr(cmd *exec.Cmd) {}

// KillTree terminates the process. On Unix the child is killed directly.
func KillTree(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}
