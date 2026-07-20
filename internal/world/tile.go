package world

type TileType uint8

const (
	TileFloor TileType = iota
	TileWall
	TileDoor
	TileVent
	TileSpawn
	TileExit
	TileConsole
	TileLocker
)

func (t TileType) Walkable() bool {
	switch t {
	case TileFloor, TileDoor, TileVent, TileSpawn, TileExit, TileLocker:
		return true
	default:
		return false
	}
}

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
	case TileLocker:
		return 'H'
	default:
		return '.'
	}
}
