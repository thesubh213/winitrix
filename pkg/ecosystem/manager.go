package ecosystem

import (
	"context"
	exec "os/exec"
	"sync"
	"time"

	"github.com/thesubh213/winitrix/pkg/runner"
)

// Package represents an installed package
type Package struct {
	ID      string
	Name    string
	Version string
}

// Result represents the result of a package manager operation
type Result struct {
	Success bool
	Message string
	Error   error
	Updated int
}

// ProgressUpdate contains progress information during package updates
type ProgressUpdate struct {
	PackageName string // Current package being updated
	Percentage  int    // Progress percentage (0-100)
	Status      string // Current status message
}

// PackageManager defines the interface for all supported package managers
type PackageManager interface {
	Name() string
	Detect(ctx context.Context) bool
	UpdateAll(ctx context.Context) Result
	UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result
	Clean(ctx context.Context) Result
	Doctor(ctx context.Context) Result
}

// Detection tuning.
const detectWorkers = 8 // bounded parallelism

// detectTimeout bounds each individual detection probe. It is a variable so
// tests can shorten it.
var detectTimeout = 5 * time.Second

// managerBinaries maps canonical manager names to the executables their probe
// shells out to. It is used purely as a fast negative prefilter: when none of
// the candidates resolve via PATH, the probe is skipped entirely. This saves
// more than a dozen process spawns (several of them PowerShell) per run on a
// typical machine.
var managerBinaries = map[string][]string{
	"Winget":            {"winget"},
	"Scoop":             {"scoop"},
	"Chocolatey":        {"choco"},
	"Npm":               {"npm"},
	"pnpm":              {"pnpm"},
	"Yarn":              {"yarn"},
	"Bun":               {"bun"},
	"Pip":               {"python", "python3", "py"},
	"pipx":              {"pipx"},
	"uv":                {"uv", "uvx"},
	"Poetry":            {"poetry"},
	"Conda":             {"conda"},
	"Cargo":             {"cargo"},
	"Rustup":            {"rustup"},
	".NET Tools":        {"dotnet"},
	"Go Tools":          {"go"},
	"Docker":            {"docker"},
	"Terraform":         {"terraform"},
	"Helm":              {"helm"},
	"WSL (Apt)":         {"wsl"},
	"VSCode Extensions": {"code"},
}

// GetAllManagers returns a list of all supported package managers
func GetAllManagers() []PackageManager {
	return []PackageManager{
		NewWingetManager(),
		NewScoopManager(),
		NewChocoManager(),
		NewNpmManager(),
		NewPnpmManager(),
		NewYarnManager(),
		NewBunManager(),
		NewPipManager(),
		NewPipxManager(),
		NewUvManager(),
		NewPoetryManager(),
		NewCondaManager(),
		NewCargoManager(),
		NewRustupManager(),
		NewDotnetManager(),
		NewGoManager(),
		NewDockerManager(),
		NewTerraformManager(),
		NewHelmManager(),
		NewWslManager(),
		NewVscodeManager(),
	}
}

// GetDetectedManagers returns only the package managers installed on the system.
func GetDetectedManagers(ctx context.Context) []PackageManager {
	return DetectManagers(ctx, GetAllManagers(), nil)
}

// GetDetectedManagersWithProgress probes all managers and reports progress.
func GetDetectedManagersWithProgress(ctx context.Context, cb func(found string, probed, total int)) []PackageManager {
	return DetectManagers(ctx, GetAllManagers(), cb)
}

// DetectFiltered applies the only/exclude filters *before* probing so filtered
// managers are never spawned at all.
func DetectFiltered(ctx context.Context, only, exclude []string, cb func(found string, probed, total int)) []PackageManager {
	return DetectManagers(ctx, FilterManagers(GetAllManagers(), only, exclude), cb)
}

// DetectManagers probes the given managers in parallel with bounded
// concurrency and a per-probe timeout, preserving the input order. The
// optional callback receives detection events for progress reporting.
func DetectManagers(ctx context.Context, managers []PackageManager, cb func(found string, probed, total int)) []PackageManager {
	total := len(managers)
	if total == 0 {
		return nil
	}

	var (
		mu       sync.Mutex
		done     int
		detected = make([]bool, total)
	)
	report := func(idx int, mgr PackageManager, ok bool) {
		mu.Lock()
		defer mu.Unlock()
		detected[idx] = ok
		done++
		if cb != nil {
			if ok {
				cb(mgr.Name(), done, total)
			} else {
				cb("", done, total)
			}
		}
	}

	sem := make(chan struct{}, detectWorkers)
	var wg sync.WaitGroup

	for i, mgr := range managers {
		if !binaryAvailable(mgr.Name()) {
			report(i, mgr, false)
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, manager PackageManager) {
			defer wg.Done()
			defer func() { <-sem }()

			probeCtx, cancel := context.WithTimeout(ctx, detectTimeout)
			ok := manager.Detect(probeCtx)
			cancel()

			report(idx, manager, ok)
		}(i, mgr)
	}

	wg.Wait()

	out := make([]PackageManager, 0, total)
	for i, mgr := range managers {
		if detected[i] {
			out = append(out, mgr)
		}
	}
	return out
}

// binaryAvailable reports whether any candidate executable for the manager
// resolves via PATH. Managers without a candidate entry are always probed.
func binaryAvailable(name string) bool {
	candidates := managerBinaries[name]
	if len(candidates) == 0 {
		return true
	}
	for _, bin := range candidates {
		if _, err := exec.LookPath(bin); err == nil {
			return true
		}
	}
	return false
}
