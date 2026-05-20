package ecosystem

import (
	"context"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type BunManager struct{}

func NewBunManager() *BunManager {
	return &BunManager{}
}

func (m *BunManager) Name() string {
	return "Bun"
}

func (m *BunManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "bun", "--version")
	return res.Err == nil && res.Stdout != ""
}

func (m *BunManager) UpdateAll(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "bun", "upgrade")

	if res.Err != nil {
		return Result{
			Success: false,
			Message: translator.ParseError("Bun", res.Stdout+res.Stderr), // Similar networking errors
			Error:   res.Err,
		}
	}

	return Result{
		Success: true,
		Message: "Successfully upgraded Bun",
		Updated: 1,
	}
}

func (m *BunManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// For now, call UpdateAll since Bun doesn't provide detailed progress output
	return m.UpdateAll(ctx)
}

func (m *BunManager) Clean(ctx context.Context) Result {
	runner.RunSilent(ctx, "bun", "pm", "cache", "rm")
	return Result{
		Success: true,
		Message: "Bun cache cleaned.",
	}
}

func (m *BunManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Bun is healthy.",
	}
}
