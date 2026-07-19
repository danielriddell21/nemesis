package world

import "strings"

// Coord is an integer grid coordinate.
type Coord struct {
	X, Y int
}

// Room is a carved rectangular room in grid space. Rooms are retained on the
// level because the simulation reasons about them: the hunter patrols between
// rooms and the director herds it from zone to zone.
type Room struct {
	X, Y, W, H int
}

// Center returns the room's centre cell.
func (r Room) Center() Coord {
	return Coord{X: r.X + r.W/2, Y: r.Y + r.H/2}
}

// Contains reports whether the cell lies inside the room.
func (r Room) Contains(c Coord) bool {
	return c.X >= r.X && c.X < r.X+r.W && c.Y >= r.Y && c.Y < r.Y+r.H
}

// Level is a generated map: a row-major grid of tiles plus the points of
// interest needed to play and to reason about it. A Level is produced
// deterministically from Seed (see Generate).
type Level struct {
	Width, Height int
	Tiles         []TileType // row-major: index = y*Width + x, len == Width*Height
	Spawn, Exit   Coord
	Rooms         []Room
	Consoles      []Coord   // wall-mounted objective consoles that unlock the exit
	VentMouths    []Coord   // vent cells that open onto a room or corridor
	Light         []float64 // per-tile brightness multiplier (1 = full)
	Flicker       []bool    // per-tile: lighting stutters (failing fixtures)
	Seed          int64
}

// newLevel allocates a Level of the given size filled entirely with walls.
// Generation then carves floors, vents and doors out of the solid mass.
func newLevel(width, height int, seed int64) *Level {
	tiles := make([]TileType, width*height)
	light := make([]float64, width*height)
	for i := range tiles {
		tiles[i] = TileWall
	}
	return &Level{
		Width:   width,
		Height:  height,
		Tiles:   tiles,
		Light:   light,
		Flicker: make([]bool, width*height),
		Seed:    seed,
	}
}

// InBounds reports whether (x, y) lies inside the grid.
func (l *Level) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < l.Width && y < l.Height
}

// At returns the tile at (x, y). Out-of-bounds reads return TileWall so callers
// can treat the world edge as solid without bounds-checking everywhere.
func (l *Level) At(x, y int) TileType {
	if !l.InBounds(x, y) {
		return TileWall
	}
	return l.Tiles[y*l.Width+x]
}

// LightAt returns the brightness multiplier at (x, y); out of bounds is dark.
func (l *Level) LightAt(x, y int) float64 {
	if !l.InBounds(x, y) || len(l.Light) == 0 {
		return 0
	}
	return l.Light[y*l.Width+x]
}

// FlickerAt reports whether the lighting at (x, y) stutters. Nil-safe for
// hand-built levels that omit the flicker layer.
func (l *Level) FlickerAt(x, y int) bool {
	if !l.InBounds(x, y) || len(l.Flicker) == 0 {
		return false
	}
	return l.Flicker[y*l.Width+x]
}

// set writes a tile at (x, y) if in bounds.
func (l *Level) set(x, y int, t TileType) {
	if l.InBounds(x, y) {
		l.Tiles[y*l.Width+x] = t
	}
}

// Solid reports whether (x, y) blocks movement and sight. Walls, consoles and
// the world edge are solid; doors are treated as solid here (the simulation
// decides when a specific door has been opened).
func (l *Level) Solid(x, y int) bool {
	switch l.At(x, y) {
	case TileWall, TileDoor, TileConsole:
		return true
	default:
		return false
	}
}

// RoomAt returns the index into Rooms of the room containing the cell, or -1
// when the cell lies in a corridor, vent or wall.
func (l *Level) RoomAt(c Coord) int {
	for i, r := range l.Rooms {
		if r.Contains(c) {
			return i
		}
	}
	return -1
}

// String renders the grid as ASCII, one row per line. Useful for tests and
// debugging.
func (l *Level) String() string {
	var b strings.Builder
	b.Grow((l.Width + 1) * l.Height)
	for y := range l.Height {
		for x := range l.Width {
			b.WriteRune(l.At(x, y).Rune())
		}
		b.WriteByte('\n')
	}
	return b.String()
}
