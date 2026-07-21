package world

import (
	"errors"
	"fmt"

	"github.com/danielriddell21/crucible/geom"
	"github.com/danielriddell21/crucible/level"
	"github.com/danielriddell21/crucible/worldgen"
)

// ErrUnreachable reports that no attempt produced a connected, playable deck.
var ErrUnreachable = errors.New("world: exhausted attempts producing a connected level")

// minDimension is the smallest deck edge; it matches crucible/level's own
// floor so a request never generates a grid too small to furnish.
const minDimension = 16

// Config parameterises deck generation.
type Config struct {
	Width, Height int

	Seed int64

	Consoles int

	Lockers int

	MaxAttempts int
}

func (c Config) normalized() Config {
	if c.Width < minDimension {
		c.Width = minDimension
	}
	if c.Height < minDimension {
		c.Height = minDimension
	}
	if c.Consoles <= 0 {
		c.Consoles = 3
	}
	if c.Consoles > 5 {
		c.Consoles = 5
	}
	if c.Lockers == 0 {
		c.Lockers = 4
	}
	if c.Lockers < 0 {
		c.Lockers = 0
	}
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 32
	}
	return c
}

// Generate digs a deck: crucible/level runs the room-and-corridor pipeline,
// carves the vent networks, and hangs the doors; nemesis's passes then mount
// the consoles, recess the lockers, and set the light moods. It retries with
// derived seeds until the layout is connected and playable.
func Generate(cfg Config) (*Level, error) {
	cfg = cfg.normalized()
	deck := &Level{}
	passes := []level.Pass{
		func(l *level.Level, _ *worldgen.RNG, rooms []Room) {
			level.CarveVents(l, rooms, level.VentConfig{})
		},
		func(l *level.Level, rng *worldgen.RNG, _ []Room) {
			level.PlaceDoors(l, rng, level.DoorConfig{})
		},
		func(l *level.Level, rng *worldgen.RNG, rooms []Room) {
			deck.Consoles = placeConsoles(l, rng, rooms, cfg.Consoles)
		},
		func(l *level.Level, rng *worldgen.RNG, rooms []Room) {
			deck.Lockers = placeLockers(l, rng, rooms, cfg.Lockers)
		},
		func(l *level.Level, rng *worldgen.RNG, rooms []Room) {
			deck.Flicker = assignLight(l, rng, rooms)
		},
	}
	validate := func(l *level.Level) bool { return viable(l, deck, cfg) }

	base, rooms, err := level.Generate(level.GenerateConfig{
		Width:  cfg.Width,
		Height: cfg.Height,
		Seed:   cfg.Seed,
	}, passes, validate)
	if err != nil {
		return nil, fmt.Errorf("%w: %dx%d seed=%d", ErrUnreachable, cfg.Width, cfg.Height, cfg.Seed)
	}
	deck.Level = base
	deck.Rooms = rooms
	return deck, nil
}

func viable(l *level.Level, deck *Level, cfg Config) bool {
	solid := func(c geom.Coord) bool { return l.Solid(c.X, c.Y) }
	if l.Exit == l.Spawn || !worldgen.Reachable(l.W, l.H, l.Spawn, l.Exit, solid) {
		return false
	}
	if len(deck.Consoles) != cfg.Consoles {
		return false
	}
	field := worldgen.FloodDist(l.W, l.H, l.Spawn, solid, nil)
	for _, c := range deck.Consoles {
		if !faceReachable(field, c) {
			return false
		}
	}
	return len(l.VentMouths) >= 2
}

func faceReachable(field worldgen.Field, console Coord) bool {
	for _, n := range worldgen.Neighbors4(console) {
		if field.At(n) >= 0 {
			return true
		}
	}
	return false
}
