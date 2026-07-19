package world

import "strings"

type Coord struct {
	X, Y int
}

type Room struct {
	X, Y, W, H int
}

func (r Room) Center() Coord {
	return Coord{X: r.X + r.W/2, Y: r.Y + r.H/2}
}

func (r Room) Contains(c Coord) bool {
	return c.X >= r.X && c.X < r.X+r.W && c.Y >= r.Y && c.Y < r.Y+r.H
}

type Level struct {
	Width, Height int
	Tiles         []TileType
	Spawn, Exit   Coord
	Rooms         []Room
	Consoles      []Coord
	VentMouths    []Coord
	Light         []float64
	Flicker       []bool
	Seed          int64
}

func newLevel(width, height int, seed int64) *Level {
	tiles := make([]TileType, width*height)
	for i := range tiles {
		tiles[i] = TileWall
	}
	return &Level{
		Width:   width,
		Height:  height,
		Tiles:   tiles,
		Light:   make([]float64, width*height),
		Flicker: make([]bool, width*height),
		Seed:    seed,
	}
}

func (l *Level) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < l.Width && y < l.Height
}

func (l *Level) At(x, y int) TileType {
	if !l.InBounds(x, y) {
		// Out-of-bounds reads act solid so callers can skip bounds checks.
		return TileWall
	}
	return l.Tiles[y*l.Width+x]
}

func (l *Level) LightAt(x, y int) float64 {
	if !l.InBounds(x, y) || len(l.Light) == 0 {
		return 0
	}
	return l.Light[y*l.Width+x]
}

func (l *Level) FlickerAt(x, y int) bool {
	if !l.InBounds(x, y) || len(l.Flicker) == 0 {
		return false
	}
	return l.Flicker[y*l.Width+x]
}

func (l *Level) set(x, y int, t TileType) {
	if l.InBounds(x, y) {
		l.Tiles[y*l.Width+x] = t
	}
}

func (l *Level) Solid(x, y int) bool {
	switch l.At(x, y) {
	case TileWall, TileDoor, TileConsole:
		return true
	default:
		return false
	}
}

func (l *Level) RoomAt(c Coord) int {
	for i, r := range l.Rooms {
		if r.Contains(c) {
			return i
		}
	}
	return -1
}

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
