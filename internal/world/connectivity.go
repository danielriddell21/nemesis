package world

type solidFn func(Coord) bool

func blocksClosed(l *Level) solidFn {
	return func(c Coord) bool { return l.Solid(c.X, c.Y) }
}

func blocksWalls(l *Level) solidFn {
	return func(c Coord) bool { return !l.At(c.X, c.Y).Walkable() }
}

func Reachable(l *Level, src, dst Coord) bool {
	return reachable(l, src, dst, blocksClosed(l))
}

func reachable(l *Level, src, dst Coord, solid solidFn) bool {
	dist := floodDist(l, src, solid)
	return l.InBounds(dst.X, dst.Y) && dist[dst.Y*l.Width+dst.X] >= 0
}

func StepsBetween(l *Level, src, dst Coord) int {
	dist := floodDist(l, src, blocksWalls(l))
	if !l.InBounds(dst.X, dst.Y) {
		return -1
	}
	return dist[dst.Y*l.Width+dst.X]
}

func distanceField(l *Level, src Coord) []int {
	return floodDist(l, src, blocksWalls(l))
}

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

func neighbors4(c Coord) [4]Coord {
	return [4]Coord{
		{X: c.X + 1, Y: c.Y},
		{X: c.X - 1, Y: c.Y},
		{X: c.X, Y: c.Y + 1},
		{X: c.X, Y: c.Y - 1},
	}
}
