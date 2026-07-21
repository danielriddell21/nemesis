package world

import "github.com/danielriddell21/crucible/worldgen"

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
	return worldgen.FloodDist(l.Width, l.Height, src, solid, nil).At(dst) >= 0
}

func StepsBetween(l *Level, src, dst Coord) int {
	return worldgen.FloodDist(l.Width, l.Height, src, blocksWalls(l), nil).At(dst)
}

func distanceField(l *Level, src Coord) []int {
	return worldgen.FloodDist(l.Width, l.Height, src, blocksWalls(l), nil).D
}

func neighbors4(c Coord) [4]Coord {
	return worldgen.Neighbors4(c)
}
