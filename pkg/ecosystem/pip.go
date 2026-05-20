package ecosystem

import (
	"context"
	"strings"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type PipManager struct{}

func NewPipManager() *PipManager {
	return &PipManager{}
}

func (m *PipManager) Name() string {
	return "Pip"
}

func (m *PipManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "python", "-m", "pip", "--version")
	if res.Err == nil && strings.Contains(res.Stdout, "pip") {
		return true
	}
	return false
}

func (m *PipManager) UpdateAll(ctx context.Context) Result {
	// Pip has no native "update all", so we just update pip itself.
	// Users usually manage packages in venvs.
	res := runner.RunSilent(ctx, "python", "-m", "pip", "install", "--upgrade", "pip")

	if res.Err != nil {
		friendlyMsg := translator.ParseError("Pip", res.Stdout+res.Stderr)

		return Result{
			Success: false,
			Message: friendlyMsg,
			Error:   res.Err,
		}
	}

	updated := 1 // we just updated pip

	return Result{
		Success: true,
		Message: "Successfully updated global pip",
		Updated: updated,
	}
}

func (m *PipManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// Pip doesn't provide detailed progress output, just call UpdateAll
	return m.UpdateAll(ctx)
}

func (m *PipManager) Clean(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "python", "-m", "pip", "cache", "purge")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: "Failed to purge pip cache.",
			Error:   res.Err,
		}
	}
	return Result{
		Success: true,
		Message: "Pip cache purged.",
	}
}

func (m *PipManager) Doctor(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "python", "-m", "pip", "check")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: "Pip dependency check found broken dependencies.",
			Error:   res.Err,
		}
	}
	return Result{
		Success: true,
		Message: "Pip dependencies are healthy.",
	}
}
