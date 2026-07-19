package sim

import "math"

const bodyRadius = 0.2

func resolveMove(w *World, pos Vec2, dx, dy float64) Vec2 {
	// Axis-by-axis resolution lets a body slide along walls instead of sticking.
	next := pos
	if !blocked(w, pos.X+dx, pos.Y) {
		next.X = pos.X + dx
	}
	if !blocked(w, next.X, pos.Y+dy) {
		next.Y = pos.Y + dy
	}
	return next
}

func blocked(w *World, x, y float64) bool {
	minX := int(math.Floor(x - bodyRadius))
	maxX := int(math.Floor(x + bodyRadius))
	minY := int(math.Floor(y - bodyRadius))
	maxY := int(math.Floor(y + bodyRadius))
	for ty := minY; ty <= maxY; ty++ {
		for tx := minX; tx <= maxX; tx++ {
			if w.Solid(tx, ty) {
				return true
			}
		}
	}
	return false
}
