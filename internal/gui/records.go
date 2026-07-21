package gui

import (
	"fmt"

	"github.com/danielriddell21/crucible/store"
)

type Records struct {
	Runs        int   `json:"runs"`
	DecksClears int   `json:"decks_cleared"`
	DeepestDeck int   `json:"deepest_deck"`
	BestDeckMS  int64 `json:"best_deck_ms,omitempty"`

	path string
}

func recordsPath() string {
	p, _ := store.Path("nemesis", "records.json")
	return p
}

func LoadRecords() Records {
	return loadRecords(recordsPath())
}

func loadRecords(path string) Records {
	r := store.Load(path, Records{})
	r.path = path
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
	if err := store.Save(r.path, r); err != nil {
		return fmt.Errorf("save records: %w", err)
	}
	return nil
}
