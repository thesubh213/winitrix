package ecosystem

import (
	"context"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type RustupManager struct{}

func NewRustupManager() *RustupManager {
	return &RustupManager{}
}

func (m *RustupManager) Name() string {
	return "Rustup"
}

func (m *RustupManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "rustup", "--version")
	return res.Err == nil && res.Stdout != ""
}

func (m *RustupManager) UpdateAll(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "rustup", "update")

	if res.Err != nil {
		return Result{
			Success: false,
			Message: translator.ParseError("Rustup", res.Stdout+res.Stderr),
			Error:   res.Err,
		}
	}

	return Result{
		Success: true,
		Message: "Successfully updated Rust toolchains",
		Updated: 1,
	}
}

func (m *RustupManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// For now, call UpdateAll since rustup doesn't provide detailed progress output
	return m.UpdateAll(ctx)
}

func (m *RustupManager) Clean(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Rustup doesn't require manual cleaning.",
	}
}

func (m *RustupManager) Doctor(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "rustup", "check")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: "Rustup toolchain check found issues.",
			Error:   res.Err,
		}
	}
	return Result{
		Success: true,
		Message: "Rustup is healthy.",
	}
}
