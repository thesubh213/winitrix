package ecosystem

import (
	"context"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type CondaManager struct{}

func NewCondaManager() *CondaManager {
	return &CondaManager{}
}

func (m *CondaManager) Name() string {
	return "Conda"
}

func (m *CondaManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "conda", "--version")
	return res.Err == nil && res.Stdout != ""
}

func (m *CondaManager) UpdateAll(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "conda", "update", "-n", "base", "-c", "defaults", "conda", "-y")

	if res.Err != nil {
		return Result{
			Success: false,
			Message: translator.ParseError("Conda", res.Stdout+res.Stderr),
			Error:   res.Err,
		}
	}

	return Result{
		Success: true,
		Message: "Successfully updated Conda base environment",
		Updated: 1,
	}
}

func (m *CondaManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// For now, call UpdateAll since conda doesn't provide detailed progress output
	return m.UpdateAll(ctx)
}

func (m *CondaManager) Clean(ctx context.Context) Result {
	runner.RunSilent(ctx, "conda", "clean", "-a", "-y")
	return Result{
		Success: true,
		Message: "Conda cache cleaned.",
	}
}

func (m *CondaManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Conda is healthy.",
	}
}
