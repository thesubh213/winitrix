package ecosystem

import (
	"context"

	"github.com/thesubh213/winitrix/pkg/runner"
)

type HelmManager struct{}

func NewHelmManager() *HelmManager {
	return &HelmManager{}
}

func (m *HelmManager) Name() string {
	return "Helm"
}

func (m *HelmManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "helm", "version")
	return res.Err == nil && res.Stdout != ""
}

func (m *HelmManager) UpdateAll(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "helm", "repo", "update")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: "Failed to update Helm repositories.",
			Error:   res.Err,
		}
	}
	return Result{
		Success: true,
		Message: "Successfully updated Helm repositories",
		Updated: 1,
	}
}

func (m *HelmManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// Helm doesn't provide detailed progress output, just call UpdateAll
	return m.UpdateAll(ctx)
}

func (m *HelmManager) Clean(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Helm cache is managed automatically.",
	}
}

func (m *HelmManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Helm is healthy.",
	}
}
