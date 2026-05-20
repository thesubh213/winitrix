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

type WingetManager struct{}

func NewWingetManager() *WingetManager {
	return &WingetManager{}
}

func (m *WingetManager) Name() string {
	return "Winget"
}

func (m *WingetManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "winget", "--version")
	if res.Err == nil && res.Stdout != "" {
		return true
	}
	return false
}

func (m *WingetManager) UpdateAll(ctx context.Context) Result {
	// --accept-source-agreements --accept-package-agreements to avoid interactive prompts
	res := runner.RunSilent(ctx, "winget", "upgrade", "--all", "--accept-source-agreements", "--accept-package-agreements")

	if res.Err != nil {
		if elevation.CheckAdminRequirement(res.Stdout + res.Stderr) {
			return Result{
				Success: false,
				Message: "Administrator privileges required to update some Winget packages.",
				Error:   res.Err,
			}
		}

		friendlyMsg := translator.ParseError("Winget", res.Stdout+res.Stderr)

		return Result{
			Success: false,
			Message: friendlyMsg,
			Error:   res.Err,
		}
	}

	// Basic heuristic to count updated packages
	updated := strings.Count(res.Stdout, "Successfully installed")

	return Result{
		Success: true,
		Message: "Successfully updated Winget packages",
		Updated: updated,
	}
}

func (m *WingetManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	logger.Debug("Starting Winget update with progress tracking")

	// Create a progress callback that parses winget output
	wrappedCallback := func(line string) {
		if progressCb != nil {
			progressCb(line)
		}

		// Parse winget output for progress information
		// Winget outputs lines like:
		// "Installing: Package Name [version]"
		// "Successfully installed Package Name [version]"
		parseWingetProgress(line)
	}

	streamRes := runner.RunWithProgress(ctx, wrappedCallback, "winget", "upgrade", "--all", "--accept-source-agreements", "--accept-package-agreements")

	if streamRes.Err != nil {
		if elevation.CheckAdminRequirement(streamRes.Stdout + streamRes.Stderr) {
			return Result{
				Success: false,
				Message: "Administrator privileges required to update some Winget packages.",
				Error:   streamRes.Err,
			}
		}

		friendlyMsg := translator.ParseError("Winget", streamRes.Stdout+streamRes.Stderr)

		return Result{
			Success: false,
			Message: friendlyMsg,
			Error:   streamRes.Err,
		}
	}

	// Count updated packages
	updated := strings.Count(streamRes.Stdout, "Successfully installed")

	return Result{
		Success: true,
		Message: "Successfully updated Winget packages",
		Updated: updated,
	}
}

// parseWingetProgress extracts package name from winget output lines
func parseWingetProgress(line string) {
	// Try to match "Installing: Package [version]" pattern
	re := regexp.MustCompile(`Installing:\s+(.+?)\s+\[.*\]`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		logger.Debug("Installing package: %s", matches[1])
	}
}

func (m *WingetManager) Clean(ctx context.Context) Result {
	return Result{
		Success: true,
		Message: "Winget does not require manual cache cleaning.",
	}
}

func (m *WingetManager) Doctor(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "winget", "source", "list")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: "Winget sources are broken.",
			Error:   res.Err,
		}
	}
	return Result{
		Success: true,
		Message: "Winget is healthy and sources are reachable.",
	}
}
