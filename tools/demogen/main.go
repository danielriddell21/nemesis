// Command demogen renders nemesis's documentation media headlessly: a
// first-person gameplay clip, an AI-visualiser clip, a station montage, and
// feature stills. Like pandemonium's generator it drives the simulation and
// the software renderers directly and never imports the GUI, so the whole set
// builds with no display.
package main

import (
	"fmt"
	"os"

	"github.com/danielriddell21/nemesis/internal/pilot"
	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/telemetry"
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
	if err := recordHunter("docs/demos/hunter.gif"); err != nil {
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

	// observe, when set, subscribes to the simulation's events; onStart, when
	// set, is called with each new station's seed. The visualiser clip needs
	// both to keep its map and feed in step; the first-person clip needs
	// neither.
	observe telemetry.Subscriber
	onStart func(seed int64)
}

func (s *session) start() error {
	seed := demoSeed + int64(s.run)*0x9e3779b9
	l, err := world.Generate(world.Config{
		Width: demoWidth, Height: demoHeight,
		Seed: seed, Consoles: demoConsoles,
	})
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	opts := []sim.Option{sim.WithLearned(s.carried), sim.WithDepth(s.run)}
	if s.observe != nil {
		opts = append(opts, sim.WithObserver(telemetry.NewBus(s.observe)))
	}
	s.game = sim.New(l, opts...)
	if s.onStart != nil {
		s.onStart(seed)
	}
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
