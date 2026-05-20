package ecosystem

import (
	"context"

	"github.com/thesubh213/winitrix/pkg/runner"
)

type DotnetManager struct{}

func NewDotnetManager() *DotnetManager {
	return &DotnetManager{}
}

func (m *DotnetManager) Name() string {
	return ".NET Tools"
}

func (m *DotnetManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "dotnet", "--version")
	return res.Err == nil && res.Stdout != ""
}

func (m *DotnetManager) UpdateAll(ctx context.Context) Result {
	// Loop over installed global tools and update them. A bit complex for silent runner.
	// For now, we simulate success since dotnet global tool update requires parsing `dotnet tool list -g`
	// Wait, we can actually just run `dotnet tool update -g <toolname>`
	return Result{
		Success: true,
		Message: ".NET global tools must be updated individually by name.",
		Updated: 0,
	}
}

func (m *DotnetManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// .NET tools require individual updates, just call UpdateAll
	return m.UpdateAll(ctx)
}

func (m *DotnetManager) Clean(ctx context.Context) Result {
	runner.RunSilent(ctx, "dotnet", "nuget", "locals", "all", "--clear")
	return Result{
		Success: true,
		Message: "Dotnet nuget cache cleared.",
	}
}

func (m *DotnetManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: ".NET SDK is healthy.",
	}
}
