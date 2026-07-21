package world

import (
	"github.com/danielriddell21/crucible/geom"
	"github.com/danielriddell21/crucible/level"
)

// Coord is a tile-grid cell and Room an axis-aligned span of tiles, shared
// with the family through crucible/geom.
type (
	Coord = geom.Coord
	Room  = geom.Rect
)

// Tile is the engine's spatial tile vocabulary from crucible/level; the
// aliases below give nemesis's two gameplay tiles their station names.
type Tile = level.Tile

const (
	TileFloor = level.TileFloor
	TileWall  = level.TileWall
	TileDoor  = level.TileDoor
	TileVent  = level.TileVent
	// TileConsole is a wall-mounted console the player brings online; the
	// engine models it as a generic wall-mounted interactable.
	TileConsole = level.TileSwitch
	// TileLocker is a recess the player ducks into to break line of sight;
	// the engine models it as a generic hiding spot.
	TileLocker = level.TileCover
	TileSpawn  = level.TileSpawn
	TileExit   = level.TileExit
)

// Level is a crucible/level spatial world plus nemesis's gameplay markers:
// the rooms, the consoles to bring online, the lockers to hide in, and the
// per-cell flicker of failing light fixtures.
type Level struct {
	*level.Level
	Rooms    []Room
	Consoles []Coord
	Lockers  []Coord
	Flicker  []bool
}

// FlickerAt reports whether the light fixture over (x, y) is a failing one
// that stutters.
func (l *Level) FlickerAt(x, y int) bool {
	if !l.InBounds(x, y) || len(l.Flicker) == 0 {
		return false
	}
	return l.Flicker[l.Index(x, y)]
}

// RoomAt returns the index of the room containing c, or -1 when c is not in
// any room.
func (l *Level) RoomAt(c Coord) int {
	for i, r := range l.Rooms {
		if r.Contains(c) {
			return i
		}
	}
	return -1
}
