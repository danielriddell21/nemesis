package gui

import (
	"path/filepath"
	"testing"
)

func TestRecordsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "records.json")
	r := loadRecords(path)
	if r.Runs != 0 || r.Escapes != 0 {
		t.Fatal("missing file should load zero records")
	}
	r.record(false, 30)
	r.record(true, 95.5)
	r.record(true, 61.2)
	if err := r.save(); err != nil {
		t.Fatal(err)
	}
	got := loadRecords(path)
	if got.Runs != 3 || got.Escapes != 2 {
		t.Errorf("got %d runs / %d escapes, want 3/2", got.Runs, got.Escapes)
	}
	if got.BestMS != 61200 {
		t.Errorf("best = %dms, want the faster escape kept", got.BestMS)
	}
}

func TestRecordsCorruptFileResets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "records.json")
	if err := writeFile(path, "{not json"); err != nil {
		t.Fatal(err)
	}
	r := loadRecords(path)
	if r.Runs != 0 {
		t.Error("corrupt records should load as zero")
	}
}

func TestRecordsEmptyPathIsInert(t *testing.T) {
	r := loadRecords("")
	r.record(true, 10)
	if err := r.save(); err != nil {
		t.Errorf("saving with no path should be a no-op, got %v", err)
	}
}
