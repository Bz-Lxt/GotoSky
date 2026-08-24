package engine_test

import (
	"testing"
	"time"

	"github.com/gotosky/gotosky/internal/engine"
)

func TestWindowsPreserveSeparatedRuns(t *testing.T) {
	t0 := time.Date(2026, time.August, 24, 13, 0, 0, 0, time.UTC)
	slots := []engine.SlotScore{
		{At: t0, Score: 70, Limit: "SEEING"},
		{At: t0.Add(time.Hour), Score: 65, Limit: "SEEING"},
		{At: t0.Add(2 * time.Hour), Score: 10, Limit: "CLOUD"},
		{At: t0.Add(3 * time.Hour), Score: 80, Limit: "WIND"},
		{At: t0.Add(4 * time.Hour), Score: 75, Limit: "WIND"},
	}
	id := engine.DefaultProfile().ID

	windows := engine.Windows(id, id, id, time.UTC, slots)
	if len(windows) != 2 {
		t.Fatalf("got %d windows, want 2: %+v", len(windows), windows)
	}

	got := make(map[time.Time]time.Time, len(windows))
	for _, window := range windows {
		got[window.StartUTC] = window.EndUTC
	}
	want := map[time.Time]time.Time{
		t0:                    t0.Add(2 * time.Hour),
		t0.Add(3 * time.Hour): t0.Add(5 * time.Hour),
	}
	for start, end := range want {
		if gotEnd, ok := got[start]; !ok || !gotEnd.Equal(end) {
			t.Errorf("window %s ended at %s (present=%v), want %s", start, gotEnd, ok, end)
		}
	}
}
