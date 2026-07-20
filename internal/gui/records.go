package gui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Records struct {
	Runs        int   `json:"runs"`
	DecksClears int   `json:"decks_cleared"`
	DeepestDeck int   `json:"deepest_deck"`
	BestDeckMS  int64 `json:"best_deck_ms,omitempty"`

	path string
}

func recordsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "nemesis", "records.json")
}

func LoadRecords() Records {
	return loadRecords(recordsPath())
}

func loadRecords(path string) Records {
	r := Records{path: path}
	if path == "" {
		return r
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return r
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return Records{path: path}
	}
	return r
}

// clearedDeck records a deck cleared at the given depth (0-based) in `elapsed`
// seconds. The deepest deck reached and the fastest single clear are kept.
func (r *Records) clearedDeck(deck int, elapsed float64) {
	r.DecksClears++
	if deck+1 > r.DeepestDeck {
		r.DeepestDeck = deck + 1
	}
	ms := int64(elapsed * 1000)
	if r.BestDeckMS == 0 || ms < r.BestDeckMS {
		r.BestDeckMS = ms
	}
}

// startedRun counts a fresh descent and remembers the deck it ended on.
func (r *Records) startedRun() {
	r.Runs++
}

func (r *Records) diedOn(deck int) {
	if deck+1 > r.DeepestDeck {
		r.DeepestDeck = deck + 1
	}
}

func (r Records) summary() []string {
	best := "--"
	if r.BestDeckMS > 0 {
		best = fmt.Sprintf("%d.%02ds", r.BestDeckMS/1000, (r.BestDeckMS%1000)/10)
	}
	return []string{
		fmt.Sprintf("RUNS %d    DECKS CLEARED %d", r.Runs, r.DecksClears),
		fmt.Sprintf("DEEPEST DECK %d    FASTEST CLEAR %s", r.DeepestDeck, best),
	}
}

func (r Records) save() error {
	if r.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o750); err != nil {
		return fmt.Errorf("records dir: %w", err)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("records encode: %w", err)
	}
	if err := os.WriteFile(r.path, data, 0o600); err != nil {
		return fmt.Errorf("records write: %w", err)
	}
	return nil
}
