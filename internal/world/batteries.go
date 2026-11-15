package world

import (
	"github.com/danielriddell21/crucible/level"
	"github.com/danielriddell21/crucible/worldgen"
)

func placeBatteries(l *level.Level, rng *worldgen.RNG, rooms []Room, want int) []Coord {
	if want <= 0 || len(rooms) == 0 {
		return nil
	}
	cells := make([]Coord, 0, want)
	minSpread := (l.W + l.H) / 8
	// Walk the rooms from a random start so the cells do not always sit in
	// the first rooms dug.
	off := rng.IntN(len(rooms))
	for i := range rooms {
		if len(cells) >= want {
			break
		}
		r := rooms[(off+i)%len(rooms)]
		if r.Contains(l.Spawn) {
			continue
		}
		if c, ok := batterySite(l, rng, r); ok && spacedFrom(cells, c, minSpread) {
			cells = append(cells, c)
		}
	}
	return cells
}

func batterySite(l *level.Level, rng *worldgen.RNG, r Room) (Coord, bool) {
	n := r.W * r.H
	off := rng.IntN(n)
	for i := range n {
		k := (off + i) % n
		c := Coord{X: r.X + k%r.W, Y: r.Y + k/r.W}
		if l.At(c.X, c.Y) == TileFloor {
			return c, true
		}
	}
	return Coord{}, false
}
