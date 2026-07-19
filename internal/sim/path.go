package sim

import (
	"container/heap"

	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	costFloor      = 1.0
	costVent       = 0.6
	costClosedDoor = 2.5
)

func stepCost(w *World, c world.Coord) (float64, bool) {
	switch w.Level.At(c.X, c.Y) {
	case world.TileVent:
		return costVent, true
	case world.TileDoor:
		if w.DoorOpen(c) {
			return costFloor, true
		}
		return costClosedDoor, true
	case world.TileWall, world.TileConsole:
		return 0, false
	default:
		return costFloor, true
	}
}

type pathNode struct {
	cell  world.Coord
	cost  float64
	index int
}

type pathQueue []*pathNode

func (q pathQueue) Len() int           { return len(q) }
func (q pathQueue) Less(i, j int) bool { return q[i].cost < q[j].cost }
func (q pathQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i]; q[i].index = i; q[j].index = j }
func (q *pathQueue) Push(x any)        { n := x.(*pathNode); n.index = len(*q); *q = append(*q, n) }
func (q *pathQueue) Pop() any          { old := *q; n := old[len(old)-1]; *q = old[:len(old)-1]; return n }

func findPath(w *World, src, dst world.Coord) []world.Coord {
	if src == dst {
		return nil
	}
	if _, ok := stepCost(w, dst); !ok {
		return nil
	}
	// A* with the Manhattan heuristic scaled by the cheapest step cost so vent
	// shortcuts stay admissible.
	heuristic := func(c world.Coord) float64 {
		return float64(absInt(c.X-dst.X)+absInt(c.Y-dst.Y)) * costVent
	}
	open := &pathQueue{}
	heap.Init(open)
	heap.Push(open, &pathNode{cell: src, cost: heuristic(src)})
	gScore := map[world.Coord]float64{src: 0}
	prev := map[world.Coord]world.Coord{}
	for open.Len() > 0 {
		cur := heap.Pop(open).(*pathNode).cell
		if cur == dst {
			return unwind(prev, src, dst)
		}
		for _, n := range [4]world.Coord{
			{X: cur.X + 1, Y: cur.Y},
			{X: cur.X - 1, Y: cur.Y},
			{X: cur.X, Y: cur.Y + 1},
			{X: cur.X, Y: cur.Y - 1},
		} {
			cost, ok := stepCost(w, n)
			if !ok {
				continue
			}
			tentative := gScore[cur] + cost
			if old, seen := gScore[n]; seen && tentative >= old {
				continue
			}
			gScore[n] = tentative
			prev[n] = cur
			heap.Push(open, &pathNode{cell: n, cost: tentative + heuristic(n)})
		}
	}
	return nil
}

func unwind(prev map[world.Coord]world.Coord, src, dst world.Coord) []world.Coord {
	var rev []world.Coord
	for c := dst; c != src; c = prev[c] {
		rev = append(rev, c)
	}
	path := make([]world.Coord, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		path = append(path, rev[i])
	}
	return path
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
