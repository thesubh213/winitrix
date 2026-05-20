package ecosystem

import (
	"context"

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

// GetDetectedManagers returns only the package managers installed on the system
func GetDetectedManagers(ctx context.Context) []PackageManager {
	all := GetAllManagers()
	var detected []PackageManager

	for _, m := range all {
		if m.Detect(ctx) {
			detected = append(detected, m)
		}
	}

	return detected
}
