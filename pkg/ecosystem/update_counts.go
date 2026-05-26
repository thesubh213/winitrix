package ecosystem

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/thesubh213/winitrix/pkg/runner"
)

type pipxListOutput struct {
	Venvs map[string]pipxListVenv `json:"venvs"`
}

type pipxListVenv struct {
	Metadata struct {
		MainPackage struct {
			Package        string `json:"package"`
			PackageVersion string `json:"package_version"`
		} `json:"main_package"`
	} `json:"metadata"`
}

func snapshotPipxVersions(ctx context.Context) (map[string]string, bool) {
	res := runner.RunSilent(ctx, "pipx", "list", "--json")
	if res.Err != nil {
		return nil, false
	}

	return parsePipxVersions(res.Stdout)
}

func parsePipxVersions(output string) (map[string]string, bool) {
	var parsed pipxListOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		return nil, false
	}

	versions := make(map[string]string, len(parsed.Venvs))
	for venvName, venv := range parsed.Venvs {
		packageName := strings.TrimSpace(venv.Metadata.MainPackage.Package)
		if packageName == "" {
			packageName = strings.TrimSpace(venvName)
		}

		packageVersion := strings.TrimSpace(venv.Metadata.MainPackage.PackageVersion)
		if packageName == "" || packageVersion == "" {
			continue
		}

		versions[packageName] = packageVersion
	}

	if len(parsed.Venvs) > 0 && len(versions) == 0 {
		return nil, false
	}

	return versions, true
}

func snapshotYarnGlobalVersions(ctx context.Context) (map[string]string, bool) {
	res := runner.RunSilent(ctx, "yarn", "global", "list", "--depth=0")
	if res.Err != nil {
		return nil, false
	}

	return parseYarnGlobalVersions(res.Stdout)
}

func parseYarnGlobalVersions(output string) (map[string]string, bool) {
	versions := make(map[string]string)
	sawPackageLine := false

	for _, rawLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "yarn "),
			strings.HasPrefix(line, "info "),
			strings.HasPrefix(line, "warning "),
			strings.HasPrefix(line, "success "),
			strings.HasPrefix(line, "Done "),
			strings.HasPrefix(line, "Done in "),
			strings.HasPrefix(line, "✨ "),
			strings.HasPrefix(line, "error "):
			continue
		}

		line = strings.TrimLeft(line, "│├└─ ")
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "@") {
			continue
		}

		index := strings.LastIndex(line, "@")
		if index <= 0 || index == len(line)-1 {
			sawPackageLine = true
			continue
		}

		name := strings.TrimSpace(line[:index])
		version := strings.TrimSpace(line[index+1:])
		if name == "" || version == "" {
			sawPackageLine = true
			continue
		}

		versions[name] = version
		sawPackageLine = true
	}

	if sawPackageLine && len(versions) == 0 {
		return nil, false
	}

	return versions, true
}

func countVersionChanges(before, after map[string]string) int {
	if len(before) == 0 || len(after) == 0 {
		return 0
	}

	updated := 0
	for name, beforeVersion := range before {
		afterVersion, ok := after[name]
		if !ok || afterVersion == beforeVersion {
			continue
		}
		updated++
	}

	return updated
}
