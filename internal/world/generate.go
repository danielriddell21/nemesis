package world

import (
	"errors"
	"fmt"
)

const minDimension = 16

var ErrUnreachable = errors.New("world: exhausted attempts producing a connected level")

type Config struct {
	Width, Height int

	Seed int64

	Consoles int

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

func faceReachable(l *Level, dist []int, console Coord) bool {
	for _, n := range neighbors4(console) {
		if l.InBounds(n.X, n.Y) && dist[n.Y*l.Width+n.X] >= 0 {
			return true
		}
	}
	return false
}
