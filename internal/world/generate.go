package world

import (
	"errors"
	"fmt"
)

// Minimum grid size that can hold a sensible partition.
const minDimension = 16

// ErrUnreachable is returned when generation cannot produce a level whose exit
// and consoles are all reachable within the attempt budget.
var ErrUnreachable = errors.New("world: exhausted attempts producing a connected level")

// Config controls level generation.
type Config struct {
	// Width and Height are the grid dimensions in tiles.
	Width, Height int
	// Seed makes generation deterministic: the same Config yields the same Level.
	Seed int64
	// Consoles is how many objective consoles the level requires before the
	// exit unlocks. Zero selects the default of 3.
	Consoles int
	// MaxAttempts bounds how many times generation retries when a candidate
	// level fails the reachability guarantee. Zero selects a sensible default.
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
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 32
	}
	return c
}

// Generate produces a level from cfg. It repeatedly builds candidates until one
// passes the guarantee — the exit and every console reachable from the spawn,
// the full console count placed, and a traversable vent network — then returns
// it. Each attempt derives a distinct but deterministic sub-seed, so a given
// Config always yields an identical Level.
func Generate(cfg Config) (*Level, error) {
	cfg = cfg.normalized()
	for attempt := range cfg.MaxAttempts {
		// Derive a per-attempt seed deterministically from the base seed.
		sub := cfg.Seed + int64(attempt)*0x100000001b3
		l := generateOnce(cfg, sub)
		l.Seed = cfg.Seed
		if viable(l, cfg) {
			return l, nil
		}
	}
	return nil, fmt.Errorf("%w: %dx%d seed=%d", ErrUnreachable, cfg.Width, cfg.Height, cfg.Seed)
}

// generateOnce builds a single candidate level: partition, carve rooms, connect
// them, then thread the vents and place doors, objectives and lighting.
func generateOnce(cfg Config, seed int64) *Level {
	l := newLevel(cfg.Width, cfg.Height, seed)
	g := newRNG(seed)

	root := &bspNode{bounds: Room{X: 1, Y: 1, W: cfg.Width - 2, H: cfg.Height - 2}}
	g.split(root, 0)
	g.carveRooms(root, l)
	g.connect(root, l)
	g.carveStubs(l)

	rooms := collectRooms(root)
	l.Rooms = rooms
	carveVents(l, rooms)
	placeDoors(l, g)
	placeSpawnAndExit(l, rooms)
	placeConsoles(l, g, rooms, cfg.Consoles)
	assignLight(l, g, rooms)
	return l
}

// viable checks a candidate against the generation guarantee: the exit distinct
// from and reachable from the spawn once doors open, every requested console
// placed with its face reachable, and a vent network with at least two mouths.
func viable(l *Level, cfg Config) bool {
	if l.Exit == l.Spawn || !reachable(l, l.Spawn, l.Exit, blocksWalls(l)) {
		return false
	}
	if len(l.Consoles) != cfg.Consoles {
		return false
	}
	dist := distanceField(l, l.Spawn)
	for _, c := range l.Consoles {
		if !faceReachable(l, dist, c) {
			return false
		}
	}
	return len(l.VentMouths) >= 2
}

// faceReachable reports whether some walkable cell adjacent to a console is
// reachable from the spawn, i.e. the player can stand in front of it.
func faceReachable(l *Level, dist []int, console Coord) bool {
	for _, n := range neighbors4(console) {
		if l.InBounds(n.X, n.Y) && dist[n.Y*l.Width+n.X] >= 0 {
			return true
		}
	}
	return false
}
