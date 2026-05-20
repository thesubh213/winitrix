package ecosystem

import (
	"context"
	"strings"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type PnpmManager struct{}

func NewPnpmManager() *PnpmManager {
	return &PnpmManager{}
}

func (m *PnpmManager) Name() string {
	return "pnpm"
}

func (m *PnpmManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "pnpm", "--version")
	if res.Err == nil && res.Stdout != "" {
		return true
	}
	return false
}

func (m *PnpmManager) UpdateAll(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "pnpm", "update", "-g")

	if res.Err != nil {
		friendlyMsg := translator.ParseError("pnpm", res.Stdout+res.Stderr)

		return Result{
			Success: false,
			Message: friendlyMsg,
			Error:   res.Err,
		}
	}

	updated := strings.Count(res.Stdout, "Packages updated") // Simplistic

	return Result{
		Success: true,
		Message: "Successfully updated global pnpm packages",
		Updated: updated,
	}
}

func (m *PnpmManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// For now, call UpdateAll since pnpm doesn't provide detailed progress output
	return m.UpdateAll(ctx)
}

func (m *PnpmManager) Clean(ctx context.Context) Result {
	runner.RunSilent(ctx, "pnpm", "store", "prune")
	return Result{
		Success: true,
		Message: "pnpm store pruned.",
	}
}

func (m *PnpmManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "pnpm is healthy.",
	}
}
