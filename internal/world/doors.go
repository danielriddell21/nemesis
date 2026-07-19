package world

// This file places doors in the corridors. A door sits in a one-cell-wide
// passage — walls either side, walkable ahead and behind — so it reads as a
// bulkhead between two spaces. Doors matter to the hunt: they block sight,
// opening one makes noise, and the hunter shoulders them open more slowly than
// the player does.

// maxDoors bounds how many doors a level gets.
const maxDoors = 8

// placeDoors converts a scattering of doorway-shaped corridor cells into
// doors. Cells are scanned in deterministic order; candidates adjacent to an
// existing door are skipped so bulkheads never stack into double doors.
func placeDoors(l *Level, g *rng) {
	placed := 0
	for y := 1; y < l.Height-1 && placed < maxDoors; y++ {
		for x := 1; x < l.Width-1 && placed < maxDoors; x++ {
			if l.At(x, y) != TileFloor || !doorway(l, x, y) {
				continue
			}
			if adjacentDoor(l, x, y) || !g.chance(0.4) {
				continue
			}
			l.set(x, y, TileDoor)
			placed++
		}
	}
}

// doorway reports whether (x, y) is shaped like a doorway: solid walls on one
// axis, walkable floor on the other.
func doorway(l *Level, x, y int) bool {
	wallsX := l.At(x-1, y) == TileWall && l.At(x+1, y) == TileWall
	wallsY := l.At(x, y-1) == TileWall && l.At(x, y+1) == TileWall
	openX := l.At(x-1, y).Walkable() && l.At(x+1, y).Walkable()
	openY := l.At(x, y-1).Walkable() && l.At(x, y+1).Walkable()
	return (wallsX && openY) || (wallsY && openX)
}

// adjacentDoor reports whether any cardinal neighbour is already a door.
func adjacentDoor(l *Level, x, y int) bool {
	for _, n := range neighbors4(Coord{X: x, Y: y}) {
		if l.At(n.X, n.Y) == TileDoor {
			return true
		}
	}
	return false
}
