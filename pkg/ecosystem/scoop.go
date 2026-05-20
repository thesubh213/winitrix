package ecosystem

import (
	"context"
	"strings"

	"github.com/thesubh213/winitrix/pkg/elevation"
	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type ScoopManager struct{}

func NewScoopManager() *ScoopManager {
	return &ScoopManager{}
}

func (m *ScoopManager) Name() string {
	return "Scoop"
}

func (m *ScoopManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "powershell", "-Command", "scoop --version")
	if res.Err == nil && strings.Contains(res.Stdout, "Scoop") {
		return true
	}
	return false
}

func (m *ScoopManager) UpdateAll(ctx context.Context) Result {
	// First update scoop itself
	runner.RunSilent(ctx, "powershell", "-Command", "scoop update")

	// Then update packages
	res := runner.RunSilent(ctx, "powershell", "-Command", "scoop update *")

	if res.Err != nil {
		if elevation.CheckAdminRequirement(res.Stdout + res.Stderr) {
			return Result{
				Success: false,
				Message: "Administrator privileges required to update some Scoop packages.",
				Error:   res.Err,
			}
		}

		friendlyMsg := translator.ParseError("Scoop", res.Stdout+res.Stderr)

		return Result{
			Success: false,
			Message: friendlyMsg,
			Error:   res.Err,
		}
	}

	updated := strings.Count(res.Stdout, "updating")

	return Result{
		Success: true,
		Message: "Successfully updated Scoop packages",
		Updated: updated,
	}
}

func (m *ScoopManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// Update scoop itself first
	runner.RunSilent(ctx, "powershell", "-Command", "scoop update")

	streamRes := runner.RunWithProgress(ctx, progressCb, "powershell", "-Command", "scoop update *")

	if streamRes.Err != nil {
		if elevation.CheckAdminRequirement(streamRes.Stdout + streamRes.Stderr) {
			return Result{
				Success: false,
				Message: "Administrator privileges required to update some Scoop packages.",
				Error:   streamRes.Err,
			}
		}

		friendlyMsg := translator.ParseError("Scoop", streamRes.Stdout+streamRes.Stderr)

		return Result{
			Success: false,
			Message: friendlyMsg,
			Error:   streamRes.Err,
		}
	}

	updated := strings.Count(streamRes.Stdout, "updating")

	return Result{
		Success: true,
		Message: "Successfully updated Scoop packages",
		Updated: updated,
	}
}

func (m *ScoopManager) Clean(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "powershell", "-Command", "scoop cleanup *")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: "Failed to clean Scoop cache.",
			Error:   res.Err,
		}
	}

	runner.RunSilent(ctx, "powershell", "-Command", "scoop cache rm *")

	return Result{
		Success: true,
		Message: "Scoop cache cleaned successfully.",
	}
}

func (m *ScoopManager) Doctor(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "powershell", "-Command", "scoop checkup")
	if strings.Contains(res.Stdout, "ERROR") {
		return Result{
			Success: false,
			Message: "Scoop checkup found issues.",
			Error:   res.Err,
		}
	}
	return Result{
		Success: true,
		Message: "Scoop is healthy.",
	}
}
