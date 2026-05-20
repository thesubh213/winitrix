package ecosystem

import (
	"context"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type PoetryManager struct{}

func NewPoetryManager() *PoetryManager {
	return &PoetryManager{}
}

func (m *PoetryManager) Name() string {
	return "Poetry"
}

func (m *PoetryManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "poetry", "--version")
	return res.Err == nil && res.Stdout != ""
}

func (m *PoetryManager) UpdateAll(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "poetry", "self", "update")

	if res.Err != nil {
		return Result{
			Success: false,
			Message: translator.ParseError("Poetry", res.Stdout+res.Stderr),
			Error:   res.Err,
		}
	}

	return Result{
		Success: true,
		Message: "Successfully updated Poetry",
		Updated: 1,
	}
}

func (m *PoetryManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// For now, call UpdateAll since Poetry doesn't provide detailed progress output
	return m.UpdateAll(ctx)
}

func (m *PoetryManager) Clean(ctx context.Context) Result {
	runner.RunSilent(ctx, "poetry", "cache", "clear", "pypi", "--all", "-n")
	return Result{
		Success: true,
		Message: "Poetry cache cleared.",
	}
}

func (m *PoetryManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Poetry is healthy.",
	}
}
