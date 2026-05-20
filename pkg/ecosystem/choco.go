package ecosystem

import (
	"context"
	"strings"

	"github.com/thesubh213/winitrix/pkg/elevation"
	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type ChocoManager struct{}

func NewChocoManager() *ChocoManager {
	return &ChocoManager{}
}

func (m *ChocoManager) Name() string {
	return "Chocolatey"
}

func (m *ChocoManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "choco", "--version")
	if res.Err == nil && res.Stdout != "" {
		return true
	}
	return false
}

func (m *ChocoManager) UpdateAll(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "choco", "upgrade", "all", "-y")

	if res.Err != nil {
		// Check for specific exit code 3010 which means success but reboot required
		if strings.Contains(res.Err.Error(), "exit status 3010") || strings.Contains(strings.ToLower(res.Stdout), "pending reboot") {
			return Result{
				Success: true,
				Message: "Chocolatey packages updated (Reboot Required)",
				Updated: 1,
			}
		}

		if elevation.CheckAdminRequirement(res.Stdout + res.Stderr) {
			return Result{
				Success: false,
				Message: "Administrator privileges required to update Choco packages.",
				Error:   res.Err,
			}
		}

		friendlyMsg := translator.ParseError("Chocolatey", res.Stdout+res.Stderr)

		return Result{
			Success: false,
			Message: friendlyMsg,
			Error:   res.Err,
		}
	}

	updated := strings.Count(res.Stdout, "upgraded")

	return Result{
		Success: true,
		Message: "Successfully updated Chocolatey packages",
		Updated: updated,
	}
}

func (m *ChocoManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	streamRes := runner.RunWithProgress(ctx, progressCb, "choco", "upgrade", "all", "-y")

	if streamRes.Err != nil {
		if strings.Contains(streamRes.Err.Error(), "exit status 3010") || strings.Contains(strings.ToLower(streamRes.Stdout), "pending reboot") {
			return Result{
				Success: true,
				Message: "Chocolatey packages updated (Reboot Required)",
				Updated: 1,
			}
		}

		if elevation.CheckAdminRequirement(streamRes.Stdout + streamRes.Stderr) {
			return Result{
				Success: false,
				Message: "Administrator privileges required to update Choco packages.",
				Error:   streamRes.Err,
			}
		}

		friendlyMsg := translator.ParseError("Chocolatey", streamRes.Stdout+streamRes.Stderr)

		return Result{
			Success: false,
			Message: friendlyMsg,
			Error:   streamRes.Err,
		}
	}

	updated := strings.Count(streamRes.Stdout, "upgraded")

	return Result{
		Success: true,
		Message: "Successfully updated Chocolatey packages",
		Updated: updated,
	}
}

func (m *ChocoManager) Clean(ctx context.Context) Result {
	// choco doesn't have a direct clean command that is safe, but we can clean the local cache manually if needed.
	// We'll simulate a success here as choco manages its own cache.
	return Result{
		Success: true,
		Message: "Chocolatey cache does not need manual cleanup.",
	}
}

func (m *ChocoManager) Doctor(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "choco", "outdated")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: "Chocolatey outdated check failed.",
			Error:   res.Err,
		}
	}
	return Result{
		Success: true,
		Message: "Chocolatey is healthy.",
	}
}
