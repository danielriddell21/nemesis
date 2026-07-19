// Package world generates and represents game levels as 2D tile grids. It is
// pure data and logic: it has no knowledge of rendering or simulation, imports
// no graphics libraries, and is fully testable headlessly. Levels are produced
// deterministically from a seed.
package world

// TileType enumerates the kinds of cell a level grid can contain.
type TileType uint8

const (
	// TileFloor is open, walkable space.
	TileFloor TileType = iota
	// TileWall is solid and blocks movement and sight.
	TileWall
	// TileDoor is a passage that blocks until opened.
	TileDoor
	// TileVent is a crawl-height duct carved through the wall mass. Both the
	// player and the hunter can traverse it; moving through one is slow and
	// creaks.
	TileVent
	// TileSpawn marks where the player starts. It is walkable.
	TileSpawn
	// TileExit marks the escape airlock floor. It is walkable; whether stepping
	// on it ends the level depends on the objectives, which the simulation
	// tracks.
	TileExit
	// TileConsole is a wall-mounted objective console the player activates with
	// use. It is solid like a wall; the consoles a level requires are recorded
	// in Level.Consoles.
	TileConsole
)

// Walkable reports whether an actor can stand on this tile type. Doors are
// considered walkable; whether a specific door is currently passable is tracked
// separately by the simulation.
func (t TileType) Walkable() bool {
	switch t {
	case TileFloor, TileDoor, TileVent, TileSpawn, TileExit:
		return true
	default:
		return false
	}
}

// Rune returns a compact character for debug/ASCII rendering of a grid.
func (t TileType) Rune() rune {
	switch t {
	case TileWall:
		return '#'
	case TileDoor:
		return '+'
	case TileVent:
		return '~'
	case TileSpawn:
		return 'S'
	case TileExit:
		return 'E'
	case TileConsole:
		return '!'
	default:
		return '.'
	}
}
