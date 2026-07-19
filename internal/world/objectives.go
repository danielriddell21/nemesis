package world

import "sort"

// This file places the run's points of interest: the spawn, the escape airlock,
// and the objective consoles that must all be activated before the airlock
// unlocks. Consoles are pushed far from the spawn and from each other so a run
// has to criss-cross the map while hunted.

// placeSpawnAndExit marks the first room's centre as spawn and the farthest
// non-vent cell reachable from it as the escape airlock. Choosing the exit from
// the spawn's reachability field maximises the journey and guarantees the exit
// is reachable by construction; the check in Generate remains a backstop.
func placeSpawnAndExit(l *Level, rooms []Room) {
	if len(rooms) == 0 {
		return
	}
	spawn := rooms[0].Center()
	dist := distanceField(l, spawn)
	exit, best := spawn, 0
	for i, d := range dist {
		c := Coord{X: i % l.Width, Y: i / l.Width}
		if d > best && l.At(c.X, c.Y) == TileFloor {
			best = d
			exit = c
		}
	}
	l.Spawn = spawn
	l.set(spawn.X, spawn.Y, TileSpawn)
	if exit != spawn {
		l.Exit = exit
		l.set(exit.X, exit.Y, TileExit)
	}
}

// placeConsoles mounts want objective consoles into room walls, favouring rooms
// far from the spawn and spread apart from one another. It returns early if the
// map cannot host that many; Generate treats a shortfall as a failed candidate.
func placeConsoles(l *Level, g *rng, rooms []Room, want int) {
	if len(rooms) == 0 || want <= 0 {
		return
	}
	dist := distanceField(l, l.Spawn)
	at := func(c Coord) int { return dist[c.Y*l.Width+c.X] }

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
	minSpread := (l.Width + l.Height) / 8
	for spread := minSpread; spread >= 0 && len(l.Consoles) < want; spread /= 2 {
		for _, ri := range order {
			if len(l.Consoles) >= want {
				break
			}
			c := rooms[ri].Center()
			if !spacedFrom(l.Consoles, c, spread) {
				continue
			}
			if m, ok := consoleSite(l, g, rooms[ri]); ok {
				l.set(m.X, m.Y, TileConsole)
				l.Consoles = append(l.Consoles, m)
			}
		}
		if spread == 0 {
			break
		}
	}
}

// spacedFrom reports whether c keeps at least spread Manhattan distance from
// every existing console.
func spacedFrom(consoles []Coord, c Coord, spread int) bool {
	for _, o := range consoles {
		if manhattan(o, c) < spread {
			return false
		}
	}
	return true
}

// consoleSite finds a wall cell on the room's perimeter to mount a console
// into: still solid wall, facing a floor cell of the room. The perimeter is
// scanned from a random offset so consoles don't all hug the top-left corner.
func consoleSite(l *Level, g *rng, r Room) (Coord, bool) {
	ring := roomPerimeter(r)
	off := g.intn(len(ring))
	for i := range ring {
		c := ring[(off+i)%len(ring)]
		if c.X < 1 || c.Y < 1 || c.X >= l.Width-1 || c.Y >= l.Height-1 {
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

// manhattan returns the L1 distance between two cells.
func manhattan(a, b Coord) int {
	return abs(a.X-b.X) + abs(a.Y-b.Y)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
