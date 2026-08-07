package gui

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/danielriddell21/crucible/hud"
	"github.com/danielriddell21/crucible/window"

	"github.com/danielriddell21/nemesis/internal/vis"
)

func Run(cfg Config) error {
	if cfg.Role == RoleVisualiser {
		return runVisualiser(cfg)
	}
	return runGame(cfg)
}

func runGame(cfg Config) error {
	seed := cfg.Seed
	if seed == 0 {
		seed = randomSeed()
	}

	settings := LoadSettings()
	overlay := hud.New()
	var aud *Audio
	if settings.Sound {
		a, err := NewAudio()
		if err != nil {
			fmt.Fprintln(os.Stderr, "audio disabled:", err)
		} else {
			aud = a
		}
	}
	game, err := NewGame(cfg, seed, overlay, aud, LoadRecords(), settings)
	if err != nil {
		return err
	}

	window.Configure(window.Options{Title: "nemesis", Width: 1280, Height: 800, MinWidth: 640, MinHeight: 400})
	if err := handleRunError(ebiten.RunGame(game)); err != nil {
		return fmt.Errorf("run game window: %w", err)
	}
	return nil
}

func runVisualiser(cfg Config) error {
	v := newVisualiser(cfg)
	w, h := vis.Size()
	window.Configure(window.Options{Title: "nemesis — hunter AI", Width: w, Height: h, MinWidth: w / 2, MinHeight: h / 2})
	if err := ebiten.RunGame(v); err != nil {
		return fmt.Errorf("run visualiser window: %w", err)
	}
	return nil
}
