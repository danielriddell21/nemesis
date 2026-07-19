package sim

import (
	"github.com/danielriddell21/nemesis/internal/world"
)

type World struct {
	Level  *world.Level
	opened map[world.Coord]bool
}

func NewWorld(l *world.Level) *World {
	return &World{Level: l, opened: make(map[world.Coord]bool)}
}

func (w *World) DoorOpen(c world.Coord) bool {
	return w.opened[c]
}

func (w *World) OpenDoor(c world.Coord) {
	if w.Level.At(c.X, c.Y) == world.TileDoor {
		w.opened[c] = true
	}
}

func (w *World) Solid(x, y int) bool {
	t := w.Level.At(x, y)
	if t == world.TileDoor {
		return !w.opened[world.Coord{X: x, Y: y}]
	}
	return !t.Walkable()
}

func (w *World) BlocksSight(x, y int) bool {
	return w.Solid(x, y)
}
