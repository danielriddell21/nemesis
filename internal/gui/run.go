package gui

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/danielriddell21/nemesis/internal/hud"
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

	overlay := hud.New()
	aud, err := NewAudio()
	if err != nil {
		fmt.Fprintln(os.Stderr, "audio disabled:", err)
		aud = nil
	}
	game, err := NewGame(cfg, seed, overlay, aud, LoadRecords())
	if err != nil {
		return err
	}

	ebiten.SetWindowTitle("nemesis")
	ebiten.SetWindowSize(1280, 800)
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	if err := handleRunError(ebiten.RunGame(game)); err != nil {
		return fmt.Errorf("run game window: %w", err)
	}
	return nil
}

func runVisualiser(cfg Config) error {
	v := newVisualiser(cfg)
	ebiten.SetWindowTitle("nemesis — hunter AI")
	ebiten.SetWindowSize(visWindowW, visWindowH)
	if err := ebiten.RunGame(v); err != nil {
		return fmt.Errorf("run visualiser window: %w", err)
	}
	return nil
}
