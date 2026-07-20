package gui

import (
	"path/filepath"
	"testing"
)

func TestSettingsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "settings.json")
	s := loadSettings(path)
	if !s.Sound || s.FOV == 0 {
		t.Fatal("missing file should load defaults")
	}
	s.Sound = false
	s.SFXVolume = 0.3
	s.Sensitivity = 2.2
	if err := s.save(); err != nil {
		t.Fatal(err)
	}
	got := loadSettings(path)
	if got.Sound {
		t.Error("sound should have persisted off")
	}
	if got.SFXVolume != 0.3 || got.Sensitivity != 2.2 {
		t.Errorf("values not persisted: %+v", got)
	}
}

func TestSettingsClamp(t *testing.T) {
	s := Settings{Sound: true, SFXVolume: 5, AmbientVolume: -1, Sensitivity: 99, FOV: 0.1}.clamped()
	if s.SFXVolume != 1 || s.AmbientVolume != 0 {
		t.Errorf("volumes not clamped: %+v", s)
	}
	if s.Sensitivity != 3 || s.FOV != 0.8 {
		t.Errorf("sensitivity/fov not clamped: %+v", s)
	}
}

func TestSettingsCorruptResetsToDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := writeFile(path, "garbage"); err != nil {
		t.Fatal(err)
	}
	s := loadSettings(path)
	if !s.Sound || s.FOV != defaultSettings().FOV {
		t.Error("corrupt settings should fall back to defaults")
	}
}

func TestBarMeter(t *testing.T) {
	if bar(0) != "----------" {
		t.Errorf("bar(0) = %q", bar(0))
	}
	if bar(1) != "##########" {
		t.Errorf("bar(1) = %q", bar(1))
	}
	if bar(0.5) != "#####-----" {
		t.Errorf("bar(0.5) = %q", bar(0.5))
	}
}
