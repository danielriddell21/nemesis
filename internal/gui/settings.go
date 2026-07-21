package gui

import (
	"fmt"

	"github.com/danielriddell21/crucible/store"
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
	p, _ := store.Path("nemesis", "settings.json")
	return p
}

func LoadSettings() Settings {
	return loadSettings(settingsPath())
}

func loadSettings(path string) Settings {
	s := store.Load(path, defaultSettings())
	s.path = path
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
	if err := store.Save(s.path, s); err != nil {
		return fmt.Errorf("save settings: %w", err)
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
