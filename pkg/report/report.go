package report

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thesubh213/winitrix/pkg/paths"
)

// BuildInfo captures build metadata for reports.
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// ManagerResult is a lightweight execution result.
type ManagerResult struct {
	Manager  string
	Success  bool
	Message  string
	Updated  int
	Err      error
	Duration time.Duration
}

// ManagerReport is the JSON-serializable report entry.
type ManagerReport struct {
	Manager    string `json:"manager"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Updated    int    `json:"updated"`
	Error      string `json:"error,omitempty"`
	Category   string `json:"category,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// SummaryReport aggregates overall stats.
type SummaryReport struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
	Updated int `json:"updated"`
}

// RunReport is the JSON output for a run.
type RunReport struct {
	Mode       string          `json:"mode"`
	Profile    string          `json:"profile,omitempty"`
	StartedAt  string          `json:"started_at"`
	EndedAt    string          `json:"ended_at"`
	DurationMs int64           `json:"duration_ms"`
	Build      BuildInfo       `json:"build"`
	Summary    SummaryReport   `json:"summary"`
	Managers   []ManagerReport `json:"managers"`
}

// BuildRunReport assembles a report from manager results.
func BuildRunReport(mode string, profile string, startedAt time.Time, endedAt time.Time, build BuildInfo, results []ManagerResult) RunReport {
	managers := make([]ManagerReport, 0, len(results))
	summary := SummaryReport{Total: len(results)}

	for _, res := range results {
		report := ManagerReport{
			Manager:    res.Manager,
			Success:    res.Success,
			Message:    res.Message,
			Updated:    res.Updated,
			Error:      errorString(res.Err),
			DurationMs: res.Duration.Milliseconds(),
		}
		if !res.Success {
			report.Category = categorizeError(res.Err, res.Message)
		}

		if res.Success {
			summary.Success++
			summary.Updated += res.Updated
		} else {
			summary.Failed++
		}

		managers = append(managers, report)
	}

	return RunReport{
		Mode:       mode,
		Profile:    profile,
		StartedAt:  startedAt.Format(time.RFC3339),
		EndedAt:    endedAt.Format(time.RFC3339),
		DurationMs: endedAt.Sub(startedAt).Milliseconds(),
		Build:      build,
		Summary:    summary,
		Managers:   managers,
	}
}

// SaveLastReport writes the report to the default cache path.
func SaveLastReport(rep RunReport) (string, error) {
	path, err := paths.DefaultReportPath()
	if err != nil {
		return "", err
	}
	return SaveReport(path, rep)
}

// SaveReport writes the report to the provided path.
func SaveReport(path string, rep RunReport) (string, error) {
	if err := paths.EnsureDir(filepath.Dir(path)); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, data, 0644)
}

// LoadReport reads and decodes a run report from the given path.
func LoadReport(path string) (RunReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RunReport{}, err
	}
	var rep RunReport
	if err := json.Unmarshal(data, &rep); err != nil {
		return RunReport{}, err
	}
	return rep, nil
}

// LoadLastReport reads the report from the default cache path.
func LoadLastReport() (RunReport, error) {
	path, err := paths.DefaultReportPath()
	if err != nil {
		return RunReport{}, err
	}
	return LoadReport(path)
}

// WriteReport writes the JSON report to stdout.
func WriteReport(rep RunReport) error {
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(append(data, '\n'))
	return err
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func categorizeError(err error, msg string) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}

	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "permission"),
		strings.Contains(lower, "administrator"),
		strings.Contains(lower, "access denied"),
		strings.Contains(lower, "elevat"):
		return "permission"
	case strings.Contains(lower, "network"),
		strings.Contains(lower, "offline"),
		strings.Contains(lower, "dns"),
		strings.Contains(lower, "connection"),
		strings.Contains(lower, "timeout"),
		strings.Contains(lower, "unreachable"):
		return "network"
	case strings.Contains(lower, "not found"),
		strings.Contains(lower, "no installed package"),
		strings.Contains(lower, "missing"):
		return "not_found"
	case strings.Contains(lower, "daemon"),
		strings.Contains(lower, "service"):
		return "service"
	case strings.Contains(lower, "corrupt"),
		strings.Contains(lower, "hash mismatch"):
		return "integrity"
	default:
		return "unknown"
	}
}
