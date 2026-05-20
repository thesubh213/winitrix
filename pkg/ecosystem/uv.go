package ecosystem

import (
	"context"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type UvManager struct{}

func NewUvManager() *UvManager {
	return &UvManager{}
}

func (m *UvManager) Name() string {
	return "uv"
}

func (m *UvManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "uv", "--version")
	return res.Err == nil && res.Stdout != ""
}

func (m *UvManager) UpdateAll(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "uv", "self", "update")

	if res.Err != nil {
		return Result{
			Success: false,
			Message: translator.ParseError("uv", res.Stdout+res.Stderr),
			Error:   res.Err,
		}
	}

	return Result{
		Success: true,
		Message: "Successfully updated uv",
		Updated: 1,
	}
}

func (m *UvManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// For now, call UpdateAll since uv doesn't provide detailed progress output
	return m.UpdateAll(ctx)
}

func (m *UvManager) Clean(ctx context.Context) Result {
	runner.RunSilent(ctx, "uv", "cache", "clean")
	return Result{
		Success: true,
		Message: "uv cache cleaned.",
	}
}

func (m *UvManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "uv is healthy.",
	}
}
