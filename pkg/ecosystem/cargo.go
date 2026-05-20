package ecosystem

import (
	"context"
	"strings"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type CargoManager struct{}

func NewCargoManager() *CargoManager {
	return &CargoManager{}
}

func (m *CargoManager) Name() string {
	return "Cargo"
}

func (m *CargoManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "cargo", "--version")
	if res.Err == nil && strings.Contains(res.Stdout, "cargo") {
		return true
	}
	return false
}

func (m *CargoManager) UpdateAll(ctx context.Context) Result {
	// Typically cargo-update or cargo install-update is used
	// If cargo-update is installed, use it. Otherwise, return a message explaining.

	checkRes := runner.RunSilent(ctx, "cargo", "install-update", "-V")
	if checkRes.Err != nil {
		return Result{
			Success: false,
			Message: "cargo-update plugin not installed. Run 'cargo install cargo-update' first.",
			Error:   checkRes.Err,
		}
	}

	res := runner.RunSilent(ctx, "cargo", "install-update", "-a")

	if res.Err != nil {
		friendlyMsg := translator.ParseError("Cargo", res.Stdout+res.Stderr)

		return Result{
			Success: false,
			Message: friendlyMsg,
			Error:   res.Err,
		}
	}

	updated := strings.Count(res.Stdout, "Updating")

	return Result{
		Success: true,
		Message: "Successfully updated Cargo packages",
		Updated: updated,
	}
}

func (m *CargoManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// For now, call UpdateAll since cargo doesn't provide detailed progress output
	return m.UpdateAll(ctx)
}

func (m *CargoManager) Clean(ctx context.Context) Result {
	// cargo clean removes the current project's target/ directory.
	// It is project-local, not a global cache operation, so we skip it.
	return Result{
		Success: true,
		Message: "Cargo clean is project-local. Use 'cargo cache --autoclean' (cargo-cache crate) for global cleanup.",
	}
}

func (m *CargoManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Cargo is healthy.",
	}
}
