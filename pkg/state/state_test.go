package state

import (
	"path/filepath"
	"testing"

	"github.com/thesubh213/winitrix/pkg/report"
)

func TestUpdateFromReportMergesHistory(t *testing.T) {
	previous := RunState{
		Managers: map[string]ManagerState{
			"Winget": {LastDurationMs: 5000, LastMessage: "old"},
			"Npm":    {LastDurationMs: 1000, LastMessage: "npm history"},
		},
	}

	rep := report.RunReport{
		EndedAt:    "2026-01-01T00:00:00Z",
		DurationMs: 7000,
		Summary:    report.SummaryReport{Total: 1, Success: 1},
		Managers: []report.ManagerReport{
			{Manager: "Winget", Success: true, DurationMs: 6500, Updated: 3, Message: "new"},
		},
	}

	got := UpdateFromReport(previous, rep)

	if got.Managers["Npm"].LastDurationMs != 1000 {
		t.Fatalf("history for managers outside the run was lost: %+v", got.Managers["Npm"])
	}
	if got.Managers["Winget"].LastDurationMs != 6500 {
		t.Fatalf("Winget duration not updated: %+v", got.Managers["Winget"])
	}
	if got.Managers["Winget"].LastUpdated != 3 {
		t.Fatalf("Winget updated count not recorded: %+v", got.Managers["Winget"])
	}
	if !got.LastSuccess {
		t.Fatal("run with zero failures must be marked successful")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "state.json")

	st := RunState{
		LastRunAt:      "2026-02-02T10:00:00Z",
		LastDurationMs: 4242,
		LastSuccess:    true,
		Managers: map[string]ManagerState{
			"Winget": {LastDurationMs: 4200, LastSuccess: true},
		},
	}

	if _, err := Save(path, st); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, _, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if got.LastDurationMs != 4242 || got.Managers["Winget"].LastDurationMs != 4200 {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}

func TestDurationMapSkipsEmptyEntries(t *testing.T) {
	st := RunState{Managers: map[string]ManagerState{
		"A": {LastDurationMs: 100},
		"B": {LastDurationMs: 0},
	}}
	got := st.DurationMap()
	if len(got) != 1 {
		t.Fatalf("DurationMap = %v, want only entries with durations", got)
	}
	if got["A"].Milliseconds() != 100 {
		t.Fatalf("A duration = %s", got["A"])
	}
}
