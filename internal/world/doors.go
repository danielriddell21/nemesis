package world

const maxDoors = 8

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

func doorway(l *Level, x, y int) bool {
	wallsX := l.At(x-1, y) == TileWall && l.At(x+1, y) == TileWall
	wallsY := l.At(x, y-1) == TileWall && l.At(x, y+1) == TileWall
	openX := l.At(x-1, y).Walkable() && l.At(x+1, y).Walkable()
	openY := l.At(x, y-1).Walkable() && l.At(x, y+1).Walkable()
	return (wallsX && openY) || (wallsY && openX)
}

func adjacentDoor(l *Level, x, y int) bool {
	for _, n := range neighbors4(Coord{X: x, Y: y}) {
		if l.At(n.X, n.Y) == TileDoor {
			return true
		}
	}
	return false
}
