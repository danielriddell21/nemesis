package gui

import (
	"path/filepath"
	"testing"
)

func TestRecordsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "records.json")
	r := loadRecords(path)
	if r.Runs != 0 || r.DecksClears != 0 {
		t.Fatal("missing file should load zero records")
	}
	r.startedRun()
	r.clearedDeck(0, 95.5) // cleared deck 1 in 95.5s
	r.clearedDeck(1, 61.2) // cleared deck 2, faster
	r.diedOn(2)            // died on deck 3
	if err := r.save(); err != nil {
		t.Fatal(err)
	}
	got := loadRecords(path)
	if got.Runs != 1 || got.DecksClears != 2 {
		t.Errorf("got %d runs / %d decks, want 1/2", got.Runs, got.DecksClears)
	}
	if got.DeepestDeck != 3 {
		t.Errorf("deepest deck = %d, want 3", got.DeepestDeck)
	}
	if got.BestDeckMS != 61200 {
		t.Errorf("best = %dms, want the faster clear kept", got.BestDeckMS)
	}
}

func TestRecordsCorruptFileResets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "records.json")
	if err := writeFile(path, "{not json"); err != nil {
		t.Fatal(err)
	}
	if loadRecords(path).Runs != 0 {
		t.Error("corrupt records should load as zero")
	}
}

func TestRecordsEmptyPathIsInert(t *testing.T) {
	r := loadRecords("")
	r.clearedDeck(0, 10)
	if err := r.save(); err != nil {
		t.Errorf("saving with no path should be a no-op, got %v", err)
	}
}

func TestRecordsSummary(t *testing.T) {
	r := Records{Runs: 4, DecksClears: 7, DeepestDeck: 3, BestDeckMS: 45120}
	lines := r.summary()
	if len(lines) != 2 {
		t.Fatalf("summary has %d lines, want 2", len(lines))
	}
	if lines[0] == "" || lines[1] == "" {
		t.Error("summary lines should not be empty")
	}
}
