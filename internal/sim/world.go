package sim

import (
	"github.com/danielriddell21/nemesis/internal/world"
)

const doorSlideTime = 0.4

type World struct {
	Level  *world.Level
	opened map[world.Coord]bool
	slide  map[world.Coord]float64
}

func NewWorld(l *world.Level) *World {
	return &World{
		Level:  l,
		opened: make(map[world.Coord]bool),
		slide:  make(map[world.Coord]float64),
	}
}

func (w *World) DoorOpen(c world.Coord) bool {
	return w.opened[c]
}

func (w *World) OpenDoor(c world.Coord) {
	if w.Level.At(c.X, c.Y) == world.TileDoor {
		w.opened[c] = true
	}
}

// DoorSlide is how far a door has retracted, 0 (shut) to 1 (fully open). Opened
// doors ease to 1 over doorSlideTime; the renderer reads this to animate them.
func (w *World) DoorSlide(c world.Coord) float64 {
	return w.slide[c]
}

func (w *World) tickDoors(dt float64) {
	for c := range w.opened {
		if w.slide[c] < 1 {
			w.slide[c] = min(1, w.slide[c]+dt/doorSlideTime)
		}
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
