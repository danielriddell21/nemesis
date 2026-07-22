// Command demogen renders nemesis's documentation media headlessly: a
// first-person gameplay clip, a station montage, and feature stills. Like
// pandemonium's generator it drives the simulation and the software renderer
// directly and never imports the GUI, so it needs no display. The AI
// visualiser demo is recorded separately from the real window (see the
// justfile's `demos` recipe).
package main

import (
	"fmt"
	"os"

	"github.com/danielriddell21/nemesis/internal/pilot"
	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	demoSeed     = 42
	demoWidth    = 48
	demoHeight   = 32
	demoConsoles = 3

	tickDT = 1.0 / 60
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "demogen:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := os.MkdirAll("docs/demos", 0o750); err != nil {
		return fmt.Errorf("demos dir: %w", err)
	}
	if err := recordCorridors("docs/demos/corridors.gif"); err != nil {
		return err
	}
	if err := stationsMontage("docs/demos/stations.png"); err != nil {
		return err
	}
	return stills()
}

// session drives one continuing run: when the pilot dies or escapes, the next
// station carries the hunter's learning forward, exactly like the real game.
type session struct {
	game    *sim.Game
	pilot   *pilot.Pilot
	carried sim.Learned
	run     int
}

func (s *session) start() error {
	l, err := world.Generate(world.Config{
		Width: demoWidth, Height: demoHeight,
		Seed: demoSeed + int64(s.run)*0x9e3779b9, Consoles: demoConsoles,
	})
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	s.game = sim.New(l, sim.WithLearned(s.carried), sim.WithDepth(s.run))
	return nil
}

func (s *session) tick() error {
	if s.game.Dead() || s.game.Escaped() {
		s.carried = s.game.Learned()
		s.run++
		return s.start()
	}
	s.game.Tick(s.pilot.Input(s.game, tickDT), tickDT)
	return nil
}
