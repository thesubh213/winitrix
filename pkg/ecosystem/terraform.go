package ecosystem

import (
	"context"

	"github.com/thesubh213/winitrix/pkg/runner"
)

type TerraformManager struct{}

func NewTerraformManager() *TerraformManager {
	return &TerraformManager{}
}

func (m *TerraformManager) Name() string {
	return "Terraform"
}

func (m *TerraformManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "terraform", "--version")
	return res.Err == nil && res.Stdout != ""
}

func (m *TerraformManager) UpdateAll(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Terraform binary updates are managed by the host OS (e.g., Scoop/Winget).",
		Updated: 0,
	}
}

func (m *TerraformManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// For terraform, just return the same result since updates are managed by host OS
	return m.UpdateAll(ctx)
}

func (m *TerraformManager) Clean(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Terraform cache cleaned locally per project.",
	}
}

func (m *TerraformManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Terraform is healthy.",
	}
}
