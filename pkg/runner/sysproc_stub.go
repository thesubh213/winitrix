//go:build !windows

package runner

import "os/exec"

func setSysProcAttr(cmd *exec.Cmd) {}
