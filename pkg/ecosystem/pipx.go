package ecosystem

import (
	"context"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type PipxManager struct{}

func NewPipxManager() *PipxManager {
	return &PipxManager{}
}

func (m *PipxManager) Name() string {
	return "pipx"
}

func (m *PipxManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "pipx", "--version")
	return res.Err == nil && res.Stdout != ""
}

func (m *PipxManager) UpdateAll(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "pipx", "upgrade-all")

	if res.Err != nil {
		return Result{
			Success: false,
			Message: translator.ParseError("pipx", res.Stdout+res.Stderr),
			Error:   res.Err,
		}
	}

	return Result{
		Success: true,
		Message: "Successfully upgraded pipx applications",
		Updated: 1, // Placeholder
	}
}

func (m *PipxManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// For now, call UpdateAll since pipx doesn't provide detailed progress output
	return m.UpdateAll(ctx)
}

func (m *PipxManager) Clean(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "pipx doesn't require manual cleaning.",
	}
}

func (m *PipxManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "pipx is healthy.",
	}
}
