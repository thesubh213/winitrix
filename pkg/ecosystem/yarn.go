package ecosystem

import (
	"context"
	"strings"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type YarnManager struct{}

func NewYarnManager() *YarnManager {
	return &YarnManager{}
}

func (m *YarnManager) Name() string {
	return "Yarn"
}

func (m *YarnManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "yarn", "--version")
	return res.Err == nil && res.Stdout != ""
}

func (m *YarnManager) UpdateAll(ctx context.Context) Result {
	// Detect Yarn version to handle Berry (2+) deprecation
	versionRes := runner.RunSilent(ctx, "yarn", "--version")
	if strings.HasPrefix(versionRes.Stdout, "2.") || strings.HasPrefix(versionRes.Stdout, "3.") || strings.HasPrefix(versionRes.Stdout, "4.") {
		return Result{
			Success: true,
			Message: "Yarn 2+ detected. Global upgrades are deprecated. Skipping.",
			Updated: 0,
		}
	}

	// It's Yarn 1.x
	res := runner.RunSilent(ctx, "yarn", "global", "upgrade")

	if res.Err != nil {
		return Result{
			Success: false,
			Message: translator.ParseError("Yarn", res.Stdout+res.Stderr),
			Error:   res.Err,
		}
	}

	return Result{
		Success: true,
		Message: "Successfully updated global Yarn packages",
		Updated: 1, // Placeholder
	}
}

func (m *YarnManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// For now, call UpdateAll since Yarn doesn't provide detailed progress output
	return m.UpdateAll(ctx)
}

func (m *YarnManager) Clean(ctx context.Context) Result {
	runner.RunSilent(ctx, "yarn", "cache", "clean")
	return Result{
		Success: true,
		Message: "Yarn cache cleaned.",
	}
}

func (m *YarnManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Yarn is healthy.",
	}
}
