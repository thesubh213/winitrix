package ecosystem

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/thesubh213/winitrix/pkg/runner"
)

// NewDemoManagers returns fake managers used by the hidden --demo flag to
// exercise every TUI state (progress, success, skip, failure, cancellation)
// without touching real packages.
func NewDemoManagers() []PackageManager {
	return []PackageManager{
		&demoManager{name: "Winget", packages: 6, stepDelay: 900 * time.Millisecond, updated: 6},
		&demoManager{name: "Scoop", packages: 3, stepDelay: 700 * time.Millisecond, updated: 3},
		&demoManager{name: "Chocolatey", packages: 2, stepDelay: 1200 * time.Millisecond, updated: 2},
		&demoManager{name: "Npm", packages: 0, stepDelay: 300 * time.Millisecond, updated: 0},
		&demoManager{
			name:        "Cargo",
			packages:    2,
			stepDelay:   900 * time.Millisecond,
			failAfter:   1,
			failMessage: "Failed to update global Cargo packages. Install 'cargo-update' and run 'cargo install-update -a'.",
			failErr:     errors.New("exit status 1"),
		},
		&demoManager{name: "Terraform", packages: 0, stepDelay: 300 * time.Millisecond, updated: 0},
	}
}

type demoManager struct {
	name        string
	packages    int
	stepDelay   time.Duration
	updated     int
	failAfter   int
	failMessage string
	failErr     error
}

func (d *demoManager) Name() string { return d.name }

func (d *demoManager) Detect(ctx context.Context) bool { return true }

func (d *demoManager) simulate(ctx context.Context, cb runner.ProgressCallback) Result {
	slug := strings.ToLower(strings.ReplaceAll(d.name, " ", "-"))

	for i := 0; i < d.packages; i++ {
		if d.failAfter > 0 && i >= d.failAfter {
			return Result{Success: false, Message: d.failMessage, Error: d.failErr}
		}

		pkg := fmt.Sprintf("%s-package-%d", slug, i+1)
		steps := 5
		for step := 1; step <= steps; step++ {
			select {
			case <-ctx.Done():
				return Result{Success: false, Message: "cancelled by user", Error: ctx.Err()}
			case <-time.After(d.stepDelay / time.Duration(steps)):
			}

			pct := step * 100 / steps
			if cb != nil {
				cb(fmt.Sprintf("Downloading: %s [%d%%]", pkg, pct))
				cb(fmt.Sprintf("Installing: %s [1.0.%d %d%%]", pkg, i, pct))
			}
		}

		if cb != nil {
			cb(fmt.Sprintf("Successfully installed %s [1.0.%d]", pkg, i))
		}
	}

	message := "Already up to date"
	if d.updated > 0 {
		message = fmt.Sprintf("Successfully updated %d packages", d.updated)
	}
	return Result{Success: true, Message: message, Updated: d.updated}
}

func (d *demoManager) UpdateAll(ctx context.Context) Result {
	return d.simulate(ctx, nil)
}

func (d *demoManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	return d.simulate(ctx, progressCb)
}

func (d *demoManager) Clean(ctx context.Context) Result {
	return Result{Success: true, Message: "Demo mode: nothing to clean."}
}

func (d *demoManager) Doctor(ctx context.Context) Result {
	return Result{Success: true, Message: "Demo mode: healthy."}
}
