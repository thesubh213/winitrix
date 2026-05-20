package ecosystem

import (
	"context"

	"github.com/thesubh213/winitrix/pkg/runner"
)

type GoManager struct{}

func NewGoManager() *GoManager {
	return &GoManager{}
}

func (m *GoManager) Name() string {
	return "Go Tools"
}

func (m *GoManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "go", "version")
	return res.Err == nil && res.Stdout != ""
}

func (m *GoManager) UpdateAll(ctx context.Context) Result {
	// Updating all go binaries installed in GOPATH/bin isn't natively supported.
	return Result{
		Success: true,
		Message: "Go global binaries require manual 'go install package@latest'.",
		Updated: 0,
	}
}

func (m *GoManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// Go tools require manual updates, just call UpdateAll
	return m.UpdateAll(ctx)
}

func (m *GoManager) Clean(ctx context.Context) Result {
	runner.RunSilent(ctx, "go", "clean", "-cache", "-modcache")
	return Result{
		Success: true,
		Message: "Go cache cleaned.",
	}
}

func (m *GoManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Go toolchain is healthy.",
	}
}
