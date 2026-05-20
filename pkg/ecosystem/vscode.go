package ecosystem

import (
	"context"
	"strings"

	"github.com/thesubh213/winitrix/pkg/runner"
)

type VscodeManager struct{}

func NewVscodeManager() *VscodeManager {
	return &VscodeManager{}
}

func (m *VscodeManager) Name() string {
	return "VSCode Extensions"
}

func (m *VscodeManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "code", "--version")
	return res.Err == nil && res.Stdout != ""
}

func (m *VscodeManager) UpdateAll(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "code", "--update-extensions")

	if res.Err != nil {
		return Result{
			Success: false,
			Message: "Failed to update VSCode extensions.",
			Error:   res.Err,
		}
	}

	updated := strings.Count(res.Stdout, "Updating extension")

	return Result{
		Success: true,
		Message: "Successfully updated VSCode extensions",
		Updated: updated,
	}
}

func (m *VscodeManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	streamRes := runner.RunWithProgress(ctx, progressCb, "code", "--update-extensions")

	if streamRes.Err != nil {
		return Result{
			Success: false,
			Message: "Failed to update VSCode extensions.",
			Error:   streamRes.Err,
		}
	}

	updated := strings.Count(streamRes.Stdout, "Updating extension")

	return Result{
		Success: true,
		Message: "Successfully updated VSCode extensions",
		Updated: updated,
	}
}

func (m *VscodeManager) Clean(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "VSCode extensions manage their own cache.",
	}
}

func (m *VscodeManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "VSCode is healthy.",
	}
}
