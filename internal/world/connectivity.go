package world

import (
	"github.com/danielriddell21/crucible/level"
	"github.com/danielriddell21/crucible/worldgen"
)

// blocksWalls is the flood-fill predicate that treats every non-walkable
// cell — walls, consoles — as impassable while letting doors through.
func blocksWalls(l *level.Level) func(Coord) bool {
	return func(c Coord) bool { return !l.At(c.X, c.Y).Walkable() }
}

// StepsBetween returns the orthogonal step distance from src to dst across
// walkable cells, or -1 when dst is unreachable.
func StepsBetween(l *Level, src, dst Coord) int {
	return worldgen.FloodDist(l.W, l.H, src, blocksWalls(l.Level), nil).At(dst)
}

func distanceField(l *level.Level, src Coord) worldgen.Field {
	return worldgen.FloodDist(l.W, l.H, src, blocksWalls(l), nil)
}

func neighbors4(c Coord) [4]Coord {
	return worldgen.Neighbors4(c)
}
