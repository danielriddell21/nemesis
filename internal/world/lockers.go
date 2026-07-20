package world

// This file places lockers: wall-backed floor recesses the player can duck
// into to break the hunter's line of sight. They are ordinary walkable tiles,
// spread across the rooms and kept clear of the spawn and the objectives so a
// run always has somewhere to hide — and somewhere the hunter can learn to
// check.

func placeLockers(l *Level, g *rng, rooms []Room, want int) {
	if want <= 0 {
		return
	}
	minSpread := (l.Width + l.Height) / 10
	for _, r := range rooms {
		if len(l.Lockers) >= want {
			return
		}
		if r.Contains(l.Spawn) {
			continue
		}
		if c, ok := lockerSite(l, g, r); ok && spacedFrom(l.Lockers, c, minSpread) {
			l.set(c.X, c.Y, TileLocker)
			l.Lockers = append(l.Lockers, c)
		}
	}
}

// lockerSite finds a floor cell inside the room that backs onto a wall, so the
// recess reads as set into the bulkhead. The interior is scanned from a random
// offset so lockers do not all hug one corner.
func lockerSite(l *Level, g *rng, r Room) (Coord, bool) {
	cells := make([]Coord, 0, r.W*r.H)
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			cells = append(cells, Coord{X: x, Y: y})
		}
	}
	off := g.intn(len(cells))
	for i := range cells {
		c := cells[(off+i)%len(cells)]
		if l.At(c.X, c.Y) != TileFloor {
			continue
		}
		if !backsOntoWall(l, c) {
			continue
		}
		return c, true
	}
	return Coord{}, false
}

// backsOntoWall reports whether a cardinal neighbour is a solid wall.
func backsOntoWall(l *Level, c Coord) bool {
	for _, n := range neighbors4(c) {
		if l.At(n.X, n.Y) == TileWall {
			return true
		}
	}
	return false
}
