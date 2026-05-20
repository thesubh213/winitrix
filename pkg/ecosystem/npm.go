package ecosystem

import (
	"context"
	"regexp"
	"strings"

	"github.com/thesubh213/winitrix/pkg/elevation"
	"github.com/thesubh213/winitrix/pkg/logger"
	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type NpmManager struct{}

func NewNpmManager() *NpmManager {
	return &NpmManager{}
}

func (m *NpmManager) Name() string {
	return "Npm"
}

func (m *NpmManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "npm", "--version")
	if res.Err == nil && res.Stdout != "" {
		return true
	}
	return false
}

func (m *NpmManager) UpdateAll(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "npm", "update", "-g")

	if res.Err != nil {
		if elevation.CheckAdminRequirement(res.Stdout + res.Stderr) {
			return Result{
				Success: false,
				Message: "Administrator privileges required to update global NPM packages.",
				Error:   res.Err,
			}
		}

		friendlyMsg := translator.ParseError("Npm", res.Stdout+res.Stderr)

		return Result{
			Success: false,
			Message: friendlyMsg,
			Error:   res.Err,
		}
	}

	updated := strings.Count(res.Stdout, "changed") + strings.Count(res.Stdout, "added")

	return Result{
		Success: true,
		Message: "Successfully updated global NPM packages",
		Updated: updated,
	}
}

func (m *NpmManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	logger.Debug("Starting npm update with progress tracking")

	// Create a progress callback that parses npm output
	wrappedCallback := func(line string) {
		if progressCb != nil {
			progressCb(line)
		}
		parseNpmProgress(line)
	}

	streamRes := runner.RunWithProgress(ctx, wrappedCallback, "npm", "update", "-g")

	if streamRes.Err != nil {
		if elevation.CheckAdminRequirement(streamRes.Stdout + streamRes.Stderr) {
			return Result{
				Success: false,
				Message: "Administrator privileges required to update global NPM packages.",
				Error:   streamRes.Err,
			}
		}

		friendlyMsg := translator.ParseError("Npm", streamRes.Stdout+streamRes.Stderr)

		return Result{
			Success: false,
			Message: friendlyMsg,
			Error:   streamRes.Err,
		}
	}

	updated := strings.Count(streamRes.Stdout, "changed") + strings.Count(streamRes.Stdout, "added")

	return Result{
		Success: true,
		Message: "Successfully updated global NPM packages",
		Updated: updated,
	}
}

// parseNpmProgress extracts package information from npm output
func parseNpmProgress(line string) {
	// Try to match package update messages
	re := regexp.MustCompile(`(\w+)\s+\d+\.\d+\.\d+ -> (\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		logger.Debug("Updating npm package: %s -> %s", matches[1], matches[2])
	}
}

func (m *NpmManager) Clean(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "npm", "cache", "clean", "--force")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: "Failed to clean NPM cache.",
			Error:   res.Err,
		}
	}

	return Result{
		Success: true,
		Message: "NPM cache cleaned successfully.",
	}
}

func (m *NpmManager) Doctor(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "npm", "doctor")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: "NPM doctor found issues. Check logs for details.",
			Error:   res.Err,
		}
	}
	return Result{
		Success: true,
		Message: "NPM is healthy.",
	}
}
