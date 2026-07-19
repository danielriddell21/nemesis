package world

// solidFn reports whether a cell blocks movement for a reachability query. It
// lets callers pose different questions of the same map — treating doors as
// solid, as open, or excluding the vent network.
type solidFn func(Coord) bool

// blocksClosed treats walls and (still-shut) doors as solid: the question "is
// the exit reachable touching nothing?". It backs the original generation
// guarantee.
func blocksClosed(l *Level) solidFn {
	return func(c Coord) bool { return l.Solid(c.X, c.Y) }
}

// blocksWalls treats every non-walkable cell (walls, console faces) as solid
// while letting doors pass: the question "is the exit reachable once doors are
// open?".
func blocksWalls(l *Level) solidFn {
	return func(c Coord) bool { return !l.At(c.X, c.Y).Walkable() }
}

// Reachable reports whether dst can be reached from src through non-solid tiles
// in the four cardinal directions, treating doors as solid. It is a
// breadth-first flood fill and is the guarantee that backs Generate: every
// level it returns has a path from spawn to exit.
func Reachable(l *Level, src, dst Coord) bool {
	return reachable(l, src, dst, blocksClosed(l))
}

// reachable is Reachable parameterised by a solidity test.
func reachable(l *Level, src, dst Coord, solid solidFn) bool {
	dist := floodDist(l, src, solid)
	return l.InBounds(dst.X, dst.Y) && dist[dst.Y*l.Width+dst.X] >= 0
}

// StepsBetween returns the shortest traversable distance from src to dst in
// tiles (doors treated as open), or -1 if unreachable.
func StepsBetween(l *Level, src, dst Coord) int {
	dist := floodDist(l, src, blocksWalls(l))
	if !l.InBounds(dst.X, dst.Y) {
		return -1
	}
	return dist[dst.Y*l.Width+dst.X]
}

// distanceField returns per-cell step distances from src with doors treated as
// open, -1 marking unreachable cells. Generation uses it to push the exit and
// the objective consoles far apart.
func distanceField(l *Level, src Coord) []int {
	return floodDist(l, src, blocksWalls(l))
}

// floodDist runs a breadth-first search from src over cells the predicate deems
// non-solid and returns per-cell step distances (row-major), with -1 for cells
// that are solid or unreachable.
func floodDist(l *Level, src Coord, solid solidFn) []int {
	dist := make([]int, l.Width*l.Height)
	for i := range dist {
		dist[i] = -1
	}
	if !l.InBounds(src.X, src.Y) || solid(src) {
		return dist
	}
	dist[src.Y*l.Width+src.X] = 0
	queue := []Coord{src}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		base := dist[c.Y*l.Width+c.X]
		for _, n := range neighbors4(c) {
			if !l.InBounds(n.X, n.Y) || solid(n) {
				continue
			}
			idx := n.Y*l.Width + n.X
			if dist[idx] != -1 {
				continue
			}
			dist[idx] = base + 1
			queue = append(queue, n)
		}
	}
	return dist
}

// neighbors4 returns the four cardinal neighbours of c.
func neighbors4(c Coord) [4]Coord {
	return [4]Coord{
		{c.X + 1, c.Y},
		{c.X - 1, c.Y},
		{c.X, c.Y + 1},
		{c.X, c.Y - 1},
	}
}
