package gui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Settings struct {
	Sound         bool    `json:"sound"`
	SFXVolume     float64 `json:"sfx_volume"`
	AmbientVolume float64 `json:"ambient_volume"`
	Sensitivity   float64 `json:"sensitivity"`
	FOV           float64 `json:"fov"`

	path string
}

func defaultSettings() Settings {
	return Settings{
		Sound:         true,
		SFXVolume:     0.8,
		AmbientVolume: 0.6,
		Sensitivity:   1.0,
		FOV:           1.152,
	}
}

func settingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "nemesis", "settings.json")
}

func LoadSettings() Settings {
	return loadSettings(settingsPath())
}

func loadSettings(path string) Settings {
	s := defaultSettings()
	s.path = path
	if path == "" {
		return s
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	if json.Unmarshal(data, &s) != nil {
		s = defaultSettings()
		s.path = path
	}
	return s.clamped()
}

func (s Settings) clamped() Settings {
	s.SFXVolume = clamp01(s.SFXVolume)
	s.AmbientVolume = clamp01(s.AmbientVolume)
	s.Sensitivity = clampRange(s.Sensitivity, 0.2, 3.0)
	s.FOV = clampRange(s.FOV, 0.8, 1.5)
	return s
}

func (s Settings) save() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o750); err != nil {
		return fmt.Errorf("settings dir: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("settings encode: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("settings write: %w", err)
	}
	return nil
}

func clamp01(v float64) float64 { return clampRange(v, 0, 1) }

func clampRange(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
