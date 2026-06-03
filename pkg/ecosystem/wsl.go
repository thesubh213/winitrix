package ecosystem

import (
	"context"
	"strings"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type WslManager struct{}

func NewWslManager() *WslManager {
	return &WslManager{}
}

func (m *WslManager) Name() string {
	return "WSL (Apt)"
}

func (m *WslManager) Detect(ctx context.Context) bool {
	// Check WSL is installed AND at least one distribution exists
	res := runner.RunSilent(ctx, "wsl", "-l", "-q")
	if res.Err != nil {
		return false
	}
	// wsl -l -q returns empty when no distros are installed
	return strings.TrimSpace(res.Stdout) != ""
}

func (m *WslManager) UpdateAll(ctx context.Context) Result {
	// Run update and upgrade in WSL
	res := runner.RunSilent(ctx, "wsl", "-e", "sudo", "apt-get", "update")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: translator.ParseError("WSL", res.Stdout+res.Stderr),
			Error:   res.Err,
		}
	}

	res = runner.RunSilent(ctx, "wsl", "-e", "sudo", "apt-get", "upgrade", "-y")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: translator.ParseError("WSL", res.Stdout+res.Stderr),
			Error:   res.Err,
		}
	}

	updated := strings.Count(res.Stdout, "upgraded,")

	return Result{
		Success: true,
		Message: "Successfully updated WSL packages",
		Updated: updated,
	}
}

func (m *WslManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// For now, call UpdateAll since WSL doesn't provide detailed progress output
	return m.UpdateAll(ctx)
}

func (m *WslManager) Clean(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "wsl", "-e", "sudo", "apt-get", "autoremove", "-y")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: "Failed to clean WSL packages.",
			Error:   res.Err,
		}
	}
	return Result{
		Success: true,
		Message: "WSL autoremove completed.",
	}
}

func (m *WslManager) Doctor(ctx context.Context) Result {
	// Check if sudo is available and doesn't require a password
	res := runner.RunSilent(ctx, "wsl", "-e", "sudo", "-n", "apt-get", "check")
	if res.Err != nil {
		combined := strings.ToLower(res.Stdout + "\n" + res.Stderr)
		if strings.Contains(combined, "password is required") || strings.Contains(combined, "passwordless") || strings.Contains(combined, "sudo") {
			return Result{
				Success: false,
				Message: "WSL apt check failed. Winitrix requires passwordless sudo setup for WSL apt-get command.",
				Error:   res.Err,
			}
		}

		// Otherwise fallback to trying without sudo to see if default user is root (e.g. Docker/WSL custom configs)
		resNoSudo := runner.RunSilent(ctx, "wsl", "-e", "apt-get", "check")
		if resNoSudo.Err == nil {
			return Result{
				Success: true,
				Message: "WSL subsystem is healthy (running as root default user).",
			}
		}

		return Result{
			Success: false,
			Message: "WSL apt check failed. Ensure apt-get is working and sudo is passwordless.",
			Error:   res.Err,
		}
	}

	return Result{
		Success: true,
		Message: "WSL subsystem is healthy and passwordless sudo is configured.",
	}
}
