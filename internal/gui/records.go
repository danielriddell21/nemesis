package gui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Records struct {
	Runs    int   `json:"runs"`
	Escapes int   `json:"escapes"`
	BestMS  int64 `json:"best_ms,omitempty"`

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

func (r *Records) record(escaped bool, elapsed float64) {
	r.Runs++
	if !escaped {
		return
	}
	r.Escapes++
	ms := int64(elapsed * 1000)
	if r.BestMS == 0 || ms < r.BestMS {
		r.BestMS = ms
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
