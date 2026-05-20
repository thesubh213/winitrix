package ecosystem

import (
	"context"
	"strings"

	"github.com/thesubh213/winitrix/pkg/runner"
	"github.com/thesubh213/winitrix/pkg/translator"
)

type DockerManager struct{}

func NewDockerManager() *DockerManager {
	return &DockerManager{}
}

func (m *DockerManager) Name() string {
	return "Docker"
}

func (m *DockerManager) Detect(ctx context.Context) bool {
	res := runner.RunSilent(ctx, "docker", "--version")
	if res.Err == nil && strings.Contains(res.Stdout, "Docker version") {
		return true
	}
	return false
}

func (m *DockerManager) UpdateAll(ctx context.Context) Result {
	// Docker doesn't "update all" natively. It relies on Desktop app or apt.
	return Result{
		Success: true,
		Message: "Docker engine updates are managed by the host system/Docker Desktop.",
		Updated: 0,
	}
}

func (m *DockerManager) UpdateAllWithProgress(ctx context.Context, progressCb runner.ProgressCallback) Result {
	// Docker updates are managed by host system, just return the result
	return m.UpdateAll(ctx)
}

func (m *DockerManager) Clean(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "docker", "system", "prune", "-f")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: translator.ParseError("Docker", res.Stdout+res.Stderr),
			Error:   res.Err,
		}
	}
	return Result{
		Success: true,
		Message: "Docker system pruned successfully.",
	}
}

func (m *DockerManager) Doctor(ctx context.Context) Result {
	res := runner.RunSilent(ctx, "docker", "info")
	if res.Err != nil {
		return Result{
			Success: false,
			Message: "Docker daemon is not running.",
			Error:   res.Err,
		}
	}
	return Result{
		Success: true,
		Message: "Docker daemon is healthy.",
	}
}
