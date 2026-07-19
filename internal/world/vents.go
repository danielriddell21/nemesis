package world

import "sort"

// This file carves the vent network: crawl-height ducts dug through the solid
// wall mass between rooms. Vents are the map's second circulation system — the
// hunter uses them as shortcuts and ambush routes, the player as a slow, creaky
// last resort — so they deliberately bypass the door-controlled corridors.

const (
	// maxVentNetworks bounds how many separate duct systems a level grows.
	maxVentNetworks = 2
	// maxVentRoomsPer bounds how many rooms one duct system may open onto.
	maxVentRoomsPer = 4
)

// ventMouth is a candidate opening: the wall cell where a room could breach
// into a wall component.
type ventMouth struct {
	room int
	at   Coord
}

// carveVents digs ducts through the wall mass. The interior wall cells are
// grouped into connected components; a component touching several rooms can
// host a duct system, so mouths are opened from those rooms into the component
// and joined with breadth-first tunnels. Everything scans in deterministic
// order, so a seed always grows the same vents.
func carveVents(l *Level, rooms []Room) {
	comp := wallComponents(l)

	// For each component, the rooms that could open a mouth into it, and where.
	mouths := map[int][]ventMouth{}
	for ri, r := range rooms {
		seen := map[int]bool{}
		for _, c := range roomPerimeter(r) {
			id, ok := comp[c]
			if !ok || seen[id] {
				continue
			}
			seen[id] = true
			mouths[id] = append(mouths[id], ventMouth{room: ri, at: c})
		}
	}

	// Rank components by how many rooms they touch, most first; ties break on
	// the smaller component id so the choice is deterministic.
	ids := make([]int, 0, len(mouths))
	for id, ms := range mouths {
		if len(ms) >= 2 {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := ids[i], ids[j]
		if len(mouths[a]) != len(mouths[b]) {
			return len(mouths[a]) > len(mouths[b])
		}
		return a < b
	})

	networks := 0
	for _, id := range ids {
		if networks >= maxVentNetworks {
			break
		}
		ms := mouths[id]
		if len(ms) > maxVentRoomsPer {
			ms = spreadMouths(ms, maxVentRoomsPer)
		}
		if digNetwork(l, comp, id, ms) {
			networks++
		}
	}

	collectVentMouths(l)
}

// spreadMouths keeps n mouths spaced evenly across the candidate list, so a
// duct that could open onto many rooms serves a spread of them rather than a
// cluster.
func spreadMouths(ms []ventMouth, n int) []ventMouth {
	out := make([]ventMouth, 0, n)
	for i := range n {
		out = append(out, ms[i*len(ms)/n])
	}
	return out
}

// digNetwork opens the mouths and joins them with tunnels inside one wall
// component, chaining mouth to mouth. It reports whether a traversable duct
// (at least two connected mouths) was dug.
func digNetwork(l *Level, comp map[Coord]int, id int, ms []ventMouth) bool {
	inComp := func(c Coord) bool {
		if v, ok := comp[c]; ok && v == id && l.At(c.X, c.Y) == TileWall {
			return true
		}
		return l.At(c.X, c.Y) == TileVent
	}
	joined := 1
	for i := 1; i < len(ms); i++ {
		path := bfsPath(l, ms[i-1].at, ms[i].at, inComp)
		if path == nil {
			continue
		}
		for _, c := range path {
			l.set(c.X, c.Y, TileVent)
		}
		joined++
	}
	return joined >= 2
}

// wallComponents labels every interior wall cell with a connected-component id
// (4-connectivity). The one-cell border ring is excluded so ducts never breach
// the world edge.
func wallComponents(l *Level) map[Coord]int {
	comp := map[Coord]int{}
	next := 0
	for y := 1; y < l.Height-1; y++ {
		for x := 1; x < l.Width-1; x++ {
			c := Coord{X: x, Y: y}
			if l.At(x, y) != TileWall {
				continue
			}
			if _, ok := comp[c]; ok {
				continue
			}
			// Flood this component.
			comp[c] = next
			queue := []Coord{c}
			for len(queue) > 0 {
				cur := queue[0]
				queue = queue[1:]
				for _, n := range neighbors4(cur) {
					if n.X < 1 || n.Y < 1 || n.X >= l.Width-1 || n.Y >= l.Height-1 {
						continue
					}
					if l.At(n.X, n.Y) != TileWall {
						continue
					}
					if _, ok := comp[n]; ok {
						continue
					}
					comp[n] = next
					queue = append(queue, n)
				}
			}
			next++
		}
	}
	return comp
}

// roomPerimeter returns the ring of cells just outside the room in a fixed
// scan order: top edge, bottom edge, left edge, right edge.
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

// bfsPath returns the shortest path from src to dst through cells the predicate
// admits, inclusive of both ends, or nil when no such path exists.
func bfsPath(l *Level, src, dst Coord, ok func(Coord) bool) []Coord {
	if !ok(src) || !ok(dst) {
		return nil
	}
	prev := map[Coord]Coord{src: src}
	queue := []Coord{src}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if c == dst {
			break
		}
		for _, n := range neighbors4(c) {
			if !l.InBounds(n.X, n.Y) || !ok(n) {
				continue
			}
			if _, seen := prev[n]; seen {
				continue
			}
			prev[n] = c
			queue = append(queue, n)
		}
	}
	if _, found := prev[dst]; !found {
		return nil
	}
	var path []Coord
	for c := dst; ; c = prev[c] {
		path = append(path, c)
		if c == prev[c] {
			break
		}
	}
	return path
}

// collectVentMouths records every vent cell that opens onto a non-vent walkable
// cell — the grates where the duct system meets the rooms and corridors, and
// where creaks give a crawler away.
func collectVentMouths(l *Level) {
	l.VentMouths = l.VentMouths[:0]
	for y := range l.Height {
		for x := range l.Width {
			if l.At(x, y) != TileVent {
				continue
			}
			for _, n := range neighbors4(Coord{X: x, Y: y}) {
				if t := l.At(n.X, n.Y); t.Walkable() && t != TileVent {
					l.VentMouths = append(l.VentMouths, Coord{X: x, Y: y})
					break
				}
			}
		}
	}
}
