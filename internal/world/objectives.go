package world

import (
	"sort"

	"github.com/danielriddell21/crucible/level"
	"github.com/danielriddell21/crucible/worldgen"
)

func placeConsoles(l *level.Level, rng *worldgen.RNG, rooms []Room, want int) []Coord {
	if len(rooms) == 0 || want <= 0 {
		return nil
	}
	dist := distanceField(l, l.Spawn)
	at := dist.At

	// Rank candidate rooms by distance from spawn, farthest first, skipping the
	// spawn room and any room the field never reached.
	order := make([]int, 0, len(rooms))
	for i, r := range rooms {
		if r.Contains(l.Spawn) || at(r.Center()) < 0 {
			continue
		}
		order = append(order, i)
	}
	sort.Slice(order, func(i, j int) bool {
		a, b := order[i], order[j]
		da, db := at(rooms[a].Center()), at(rooms[b].Center())
		if da != db {
			return da > db
		}
		return a < b
	})

	// Greedily pick rooms that keep their pairwise spread, relaxing the spread
	// requirement if the map is too small to satisfy it.
	consoles := make([]Coord, 0, want)
	minSpread := (l.W + l.H) / 8
	for spread := minSpread; spread >= 0 && len(consoles) < want; spread /= 2 {
		consoles = mountConsoles(l, rng, rooms, order, consoles, want, spread)
		if spread == 0 {
			break
		}
	}
	return consoles
}

func mountConsoles(l *level.Level, rng *worldgen.RNG, rooms []Room, order []int, consoles []Coord, want, spread int) []Coord {
	for _, ri := range order {
		if len(consoles) >= want {
			return consoles
		}
		if !spacedFrom(consoles, rooms[ri].Center(), spread) {
			continue
		}
		if m, ok := consoleSite(l, rng, rooms[ri]); ok {
			l.Set(m.X, m.Y, TileConsole)
			consoles = append(consoles, m)
		}
	}
	return consoles
}

func spacedFrom(placed []Coord, c Coord, spread int) bool {
	for _, o := range placed {
		if manhattan(o, c) < spread {
			return false
		}
	}
	return true
}

func consoleSite(l *level.Level, rng *worldgen.RNG, r Room) (Coord, bool) {
	ring := roomPerimeter(r)
	off := rng.IntN(len(ring))
	for i := range ring {
		c := ring[(off+i)%len(ring)]
		if c.X < 1 || c.Y < 1 || c.X >= l.W-1 || c.Y >= l.H-1 {
			continue
		}
		if l.At(c.X, c.Y) != TileWall {
			continue
		}
		for _, n := range neighbors4(c) {
			if r.Contains(n) && l.At(n.X, n.Y) == TileFloor {
				return c, true
			}
		}
	}
	return Coord{}, false
}

// roomPerimeter lists the wall ring one cell outside the room.
func roomPerimeter(r Room) []Coord {
	out := make([]Coord, 0, 2*r.W+2*r.H)
	for x := r.X; x < r.X+r.W; x++ {
		out = append(out, Coord{X: x, Y: r.Y - 1})
	}
	for x := r.X; x < r.X+r.W; x++ {
		out = append(out, Coord{X: x, Y: r.Y + r.H})
	}
	for y := r.Y; y < r.Y+r.H; y++ {
		out = append(out, Coord{X: r.X - 1, Y: y})
	}
	for y := r.Y; y < r.Y+r.H; y++ {
		out = append(out, Coord{X: r.X + r.W, Y: y})
	}
	return out
}

func manhattan(a, b Coord) int {
	return abs(a.X-b.X) + abs(a.Y-b.Y)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
