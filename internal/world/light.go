package world

import (
	"github.com/danielriddell21/crucible/level"
	"github.com/danielriddell21/crucible/worldgen"
)

const (
	corridorLight = 0.4
	ventLight     = 0.15
	flickerChance = 0.18
)

// assignLight paints each room a steady base glow — some rooms flickering —
// and drops corridors and vents to their dim constants. It returns the
// per-cell flicker mask; the light itself lives on the level. crucible/level
// starts every cell fully lit, so the moods are laid over a cleared field.
func assignLight(l *level.Level, rng *worldgen.RNG, rooms []Room) []bool {
	flicker := make([]bool, l.W*l.H)
	for i := range l.Light {
		l.Light[i] = 0
	}
	for _, r := range rooms {
		base := rng.BetweenF(0.5, 0.9)
		flick := rng.Chance(flickerChance)
		for y := r.Y; y < r.Y+r.H; y++ {
			for x := r.X; x < r.X+r.W; x++ {
				i := l.Index(x, y)
				l.Light[i] = base
				flicker[i] = flick
			}
		}
	}
	for i, t := range l.Tiles {
		if l.Light[i] > 0 {
			continue
		}
		switch t {
		case TileVent:
			l.Light[i] = ventLight
		case TileFloor, TileDoor, TileSpawn, TileExit, TileLocker:
			l.Light[i] = corridorLight
		}
	}
	return flicker
}
