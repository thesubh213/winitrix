package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/thesubh213/winitrix/pkg/paths"
	"github.com/thesubh213/winitrix/pkg/report"
)

// RunState stores recent run history for ETA and reporting.
type RunState struct {
	LastRunAt      string                  `json:"last_run_at,omitempty"`
	LastDurationMs int64                   `json:"last_duration_ms,omitempty"`
	LastSuccess    bool                    `json:"last_success"`
	Managers       map[string]ManagerState `json:"managers,omitempty"`
}

// ManagerState stores recent per-manager history.
type ManagerState struct {
	LastRunAt      string `json:"last_run_at,omitempty"`
	LastSuccess    bool   `json:"last_success"`
	LastUpdated    int    `json:"last_updated,omitempty"`
	LastMessage    string `json:"last_message,omitempty"`
	LastError      string `json:"last_error,omitempty"`
	LastCategory   string `json:"last_category,omitempty"`
	LastDurationMs int64  `json:"last_duration_ms,omitempty"`
}

// Load reads run state from disk.
func Load(path string) (RunState, string, error) {
	resolved := path
	if resolved == "" {
		var err error
		resolved, err = paths.DefaultStatePath()
		if err != nil {
			return RunState{}, "", err
		}
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return RunState{}, resolved, nil
		}
		return RunState{}, resolved, err
	}

	var st RunState
	if err := json.Unmarshal(data, &st); err != nil {
		return RunState{}, resolved, err
	}

	return st, resolved, nil
}

// Save writes the state to disk.
func Save(path string, st RunState) (string, error) {
	resolved := path
	if resolved == "" {
		var err error
		resolved, err = paths.DefaultStatePath()
		if err != nil {
			return "", err
		}
	}

	if err := paths.EnsureDir(filepath.Dir(resolved)); err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return "", err
	}

	return resolved, os.WriteFile(resolved, data, 0644)
}

// UpdateFromReport builds a new RunState based on the latest report.
func UpdateFromReport(rep report.RunReport) RunState {
	st := RunState{
		LastRunAt:      rep.EndedAt,
		LastDurationMs: rep.DurationMs,
		LastSuccess:    rep.Summary.Failed == 0,
		Managers:       make(map[string]ManagerState),
	}

	for _, mgr := range rep.Managers {
		st.Managers[mgr.Manager] = ManagerState{
			LastRunAt:      rep.EndedAt,
			LastSuccess:    mgr.Success,
			LastUpdated:    mgr.Updated,
			LastMessage:    mgr.Message,
			LastError:      mgr.Error,
			LastCategory:   mgr.Category,
			LastDurationMs: mgr.DurationMs,
		}
	}

	return st
}

// DurationMap returns previous durations keyed by manager name.
func (st RunState) DurationMap() map[string]time.Duration {
	out := make(map[string]time.Duration)
	for name, mgr := range st.Managers {
		if mgr.LastDurationMs > 0 {
			out[name] = time.Duration(mgr.LastDurationMs) * time.Millisecond
		}
	}
	return out
}
