package ecosystem

import (
	"context"
	"fmt"
	"strings"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type YarnManager struct {
	// version is cached from Detect (which already runs `yarn --version`) to
	// avoid an extra process spawn per update.
	version string
}

func NewYarnManager() *YarnManager {
	return &YarnManager{}
}

func (m *YarnManager) Name() string {
	return "Yarn"
}

func (m *YarnManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "yarn", "--version")
	if res.Err != nil || strings.TrimSpace(res.Stdout) == "" {
		return false
	}
	m.version = strings.TrimSpace(res.Stdout)
	return true
}

func (m *YarnManager) yarnVersion(ctx context.Context) string {
	if m.version != "" {
		return m.version
	}
	res := runner.RunSilent(ctx, "yarn", "--version")
	m.version = strings.TrimSpace(res.Stdout)
	return m.version
}

func (m *YarnManager) UpdateAll(ctx context.Context) Result {
	// Detect Yarn version to handle Berry (2+) deprecation
	version := m.yarnVersion(ctx)
	if strings.HasPrefix(version, "2.") || strings.HasPrefix(version, "3.") || strings.HasPrefix(version, "4.") {
		return Result{
			Success: true,
			Message: "Yarn 2+ detected. Global upgrades are deprecated. Skipping.",
			Updated: 0,
		}
	}

	beforeVersions, _ := snapshotYarnGlobalVersions(ctx)

	// It's Yarn 1.x
	res := runner.RunSilent(ctx, "yarn", "global", "upgrade")

	if res.Err != nil {
		return Result{
			Success: false,
			Message: translator.ParseError("Yarn", res.Stdout+res.Stderr),
			Error:   res.Err,
		}
	}

	afterVersions, ok := snapshotYarnGlobalVersions(ctx)
	updated := 0
	if ok {
		updated = countVersionChanges(beforeVersions, afterVersions)
	}

	message := "Yarn global upgrade completed."
	if updated > 0 {
		message = fmt.Sprintf("Successfully updated %d global Yarn packages", updated)
	}

	return Result{
		Success: true,
		Message: message,
		Updated: updated,
	}
}

func (m *YarnManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// For now, call UpdateAll since Yarn doesn't provide detailed progress output
	return m.UpdateAll(ctx)
}

func (m *YarnManager) Clean(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "yarn", "cache", "clean")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: "Failed to clean Yarn cache.",
			Error:   res.Err,
		}
	}
	return Result{
		Success: true,
		Message: "Yarn cache cleaned.",
	}
}

func (m *YarnManager) Doctor(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Yarn is healthy.",
	}
}
